// src/app/program/[id]/page.tsx
'use client'

import { useEffect, useMemo, useState } from 'react'
import Link from 'next/link'
import { useParams } from 'next/navigation'
import { api } from '@/shared/api/client'
import { getIdentity } from '@/shared/dev/auth'
import { Markdown } from '@/shared/ui/Markdown'

type Program = {
  ID: string
  Title: string
  Description: string
  Status?: string
}

type Group = {
  ID: string
  Title: string
  Capacity: number
  IsOpen: boolean
  RequiresInterview: boolean
}

type ProgramWithGroups = {
  Program: Program
  Groups: Group[]
}

type AppLite = { ID: string; GroupID: string; Status: string }

export default function ProgramPage() {
  const params = useParams<{ id: string }>()
  const id = params?.id || ''

  const ident = getIdentity()

  const [data, setData] = useState<ProgramWithGroups | null>(null)
  const [err, setErr] = useState('')

  // ❗️ВАЖНО: null = еще не знаем, учитель или нет
  const [teacherMode, setTeacherMode] = useState<boolean | null>(null)

  const [myApps, setMyApps] = useState<AppLite[]>([])
  const [appsLoaded, setAppsLoaded] = useState(false)

  const [draftByGroup, setDraftByGroup] = useState<Record<string, string>>({})
  const [learnOkByGroup, setLearnOkByGroup] = useState<Record<string, boolean>>({})

  const appByGroup = useMemo(() => {
    const m = new Map<string, string>()
    for (const a of myApps) if (!m.has(a.GroupID)) m.set(a.GroupID, a.Status)
    return m
  }, [myApps])

  async function reload() {
    if (!id) return
    setErr('')
    setAppsLoaded(false)
    setTeacherMode(null)

    try {
      const pg = await api.getProgram(id)
      setData(pg)

      // teacher mode (по назначению, НЕ по роли)
      try {
        const res = await api.teacherProgramAccess(id)
        setTeacherMode(Boolean(res?.ok))
      } catch {
        setTeacherMode(false)
      }

      // заявки ученика (у тебя role всегда user, и у препода тоже user,
      // поэтому всё равно грузим — но важен факт teacherMode)
      const apps = await api.listMyApplications()
      setMyApps((apps ?? []).map((a: any) => ({ ID: a.ID, GroupID: a.GroupID, Status: a.Status })))
      setAppsLoaded(true)
    } catch (e: any) {
      setErr(e?.message || String(e))
      // даже если ошибка — помечаем как “загрузка завершена”, чтобы UI не висел
      setTeacherMode(false)
      setAppsLoaded(true)
    }
  }

  useEffect(() => {
    setData(null)
    setMyApps([])
    setDraftByGroup({})
    setLearnOkByGroup({})
    reload()
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [id])

  useEffect(() => {
    if (!appsLoaded) return
    if (!data) return

    for (const g of data.Groups ?? []) {
      const st = appByGroup.get(g.ID)
      if (st !== 'approved') continue
      if (learnOkByGroup[g.ID] !== undefined) continue

      api.listMaterials(g.ID)
        .then(() => setLearnOkByGroup((prev) => ({ ...prev, [g.ID]: true })))
        .catch(() => setLearnOkByGroup((prev) => ({ ...prev, [g.ID]: false })))
    }
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [data, appsLoaded, myApps])

  async function apply(groupId: string) {
    const st = appByGroup.get(groupId)
    if (st && st !== 'cancelled') return

    const comment = (draftByGroup[groupId] ?? '').trim()
    try {
      await api.createApplication(groupId, comment)
      setDraftByGroup((prev) => {
        const next = { ...prev }
        delete next[groupId]
        return next
      })
      await reload()
    } catch (e: any) {
      alert(e?.message || String(e))
    }
  }

  async function cancel(appId: string) {
    if (!confirm('Отменить заявку?')) return
    try {
      await api.cancelMyApplication(appId)
      await reload()
    } catch (e: any) {
      alert(e?.message || String(e))
    }
  }

  if (!data) return <div style={{ padding: 12 }}>{err || 'Загрузка...'}</div>

  // ✅ можно ли вообще показывать блок “подать заявку”
  // ВАЖНО: teacherMode должен быть известен, и мои заявки должны быть загружены,
  // иначе будет фликер.
  const readyForApplyUI = appsLoaded && teacherMode !== null

  const canApply = readyForApplyUI && ident.role === 'user' && teacherMode === false

  return (
    <div style={{ padding: 12 }}>
      <div style={{ marginBottom: 8 }}>
        <Link href="/catalog">← Каталог</Link>
      </div>

      <h2>{data.Program.Title}</h2>
      <div style={{ whiteSpace: 'pre-wrap' }}>
        <Markdown text={data.Program.Description || ''} />
      </div>

      {/* ✅ ссылка на заявки курса (только преподавателю) */}
      {teacherMode === true ? (
        <div style={{ marginTop: 8 }}>
          <Link href={`/program/${id}/applications`}>Заявки на курс</Link>
        </div>
      ) : null}

      <h3 style={{ marginTop: 12 }}>Группы</h3>

      {data.Groups.length === 0 ? (
        <div>Пока нет групп.</div>
      ) : (
        <ul>
          {data.Groups.map((g) => {
            const status = appByGroup.get(g.ID)

            return (
              <li key={g.ID} style={{ marginBottom: 14 }}>
                <div><b>{g.Title}</b></div>
                <div>
                  Мест: {g.Capacity} {g.RequiresInterview ? '• интервью' : ''}
                </div>

                {/* ✅ пока не готовы данные — НЕ рисуем форму */}
                {!readyForApplyUI ? (
                  <div style={{ marginTop: 8, opacity: 0.75 }}>Проверяю доступ...</div>
                ) : canApply ? (
                  <div style={{ marginTop: 8 }}>
                    {status ? (
                      <div>
                        <div>Заявка уже есть: {status}</div>

                        {(status === 'submitted' || status === 'in_review') ? (
                          <button
                            onClick={() => {
                              const app = myApps.find((x) => x.GroupID === g.ID)
                              if (!app) return alert('Не найдена заявка для этой группы')
                              cancel(app.ID)
                            }}
                          >
                            Отменить заявку
                          </button>
                        ) : null}

                        {status === 'cancelled' ? (
                          <div style={{ marginTop: 8 }}>
                            <div>О себе</div>
                            <textarea
                              rows={4}
                              value={draftByGroup[g.ID] ?? ''}
                              onChange={(e) =>
                                setDraftByGroup((prev) => ({ ...prev, [g.ID]: e.target.value }))
                              }
                            />
                            <div>
                              <button onClick={() => apply(g.ID)}>Подать заявку снова</button>
                            </div>
                          </div>
                        ) : null}
                      </div>
                    ) : (
                      <div>
                        <div>О себе</div>
                        <textarea
                          rows={4}
                          value={draftByGroup[g.ID] ?? ''}
                          onChange={(e) =>
                            setDraftByGroup((prev) => ({ ...prev, [g.ID]: e.target.value }))
                          }
                        />
                        <div>
                          <button onClick={() => apply(g.ID)}>Подать заявку</button>
                        </div>
                      </div>
                    )}
                  </div>
                ) : (
                  <div style={{ marginTop: 8, opacity: 0.75 }}>
                    {teacherMode === true
                      ? 'Вы в режиме преподавателя: подавать на свой курс нельзя.'
                      : status
                        ? 'Заявка уже есть — см. статус выше.'
                        : null}
                  </div>
                )}

                {/* ✅ переход в обучение */}
                {status === 'approved' && learnOkByGroup[g.ID] === true ? (
                  <div style={{ marginTop: 6 }}>
                    <Link href={`/learn/group/${g.ID}`}>Перейти в обучение</Link>
                  </div>
                ) : null}
              </li>
            )
          })}
        </ul>
      )}
    </div>
  )
}
