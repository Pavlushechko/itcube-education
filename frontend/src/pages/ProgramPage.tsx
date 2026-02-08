// src/pages/ProgramPage.tsx

import { useEffect, useMemo, useState } from 'react'
import { api } from '../lib/api'
import { useParams, Link } from 'react-router-dom'
import type { ProgramWithGroups } from '../lib/types'
import { getIdentity } from '../lib/auth'
import { AdminPublishProgram } from '../ui/AdminPublishProgram'
import { Markdown } from '../ui/Markdown'
import { GroupTeacherAdminPanel } from '../ui/GroupTeacherAdminPanel'
import { AdminCreateGroup } from '../ui/AdminCreateGroup'

type AppLite = {
  GroupID: string
  Status: string
}

export function ProgramPage() {
  const { id } = useParams()
  const [data, setData] = useState<ProgramWithGroups | null>(null)
  const [myApps, setMyApps] = useState<AppLite[]>([])
  const [err, setErr] = useState<string>('')

  const ident = getIdentity()
  const isStaff = ident.role === 'admin' || ident.role === 'moderator'
  const isPublished = data?.Program?.Status === 'published'

  const [editing, setEditing] = useState<Record<string, boolean>>({})
  const [draft, setDraft] = useState<
    Record<string, { title: string; capacity: number; is_open: boolean; requires_interview: boolean }>
  >({})
  const [saving, setSaving] = useState<Record<string, boolean>>({})

  function startEdit(g: any) {
    setEditing((m) => ({ ...m, [g.ID]: true }))
    setDraft((m) => ({
      ...m,
      [g.ID]: {
        title: g.Title ?? '',
        capacity: Number(g.Capacity ?? 0),
        is_open: Boolean(g.IsOpen ?? true),
        requires_interview: Boolean(g.RequiresInterview ?? false),
      },
    }))
  }

  function cancelEdit(groupId: string) {
    setEditing((m) => ({ ...m, [groupId]: false }))
    setDraft((m) => {
      const copy = { ...m }
      delete copy[groupId]
      return copy
    })
  }

  async function reload() {
    if (!id) return
    setErr('')

    const loadProgram = isStaff ? api.getProgramAdmin(id) : api.getProgram(id)

    try {
      const [pg, apps] = await Promise.all([loadProgram, api.listMyApplications()])
      setData(pg)

      const lite: AppLite[] = (apps ?? []).map((a: any) => ({
        GroupID: a.GroupID,
        Status: a.Status,
      }))
      setMyApps(lite)
    } catch (e: any) {
      setErr(String(e.message || e))
    }
  }

  useEffect(() => {
    setData(null)
    reload()
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [id, isStaff])

  const appByGroup = useMemo(() => {
    const m = new Map<string, string>()
    for (const a of myApps) m.set(a.GroupID, a.Status)
    return m
  }, [myApps])

  async function apply(groupId: string) {
    if (appByGroup.has(groupId)) return

    try {
      const res = await api.createApplication(groupId, 'хочу учиться')
      alert('Заявка создана: ' + res.id)
      await reload()
    } catch (e: any) {
      const msg = e?.message || String(e)

      if (typeof msg === 'string' && msg.toLowerCase().includes('duplicate')) {
        alert('Заявка на эту группу уже существует. Проверь "Мои заявки".')
        await reload()
        return
      }

      alert('Ошибка: ' + msg)
    }
  }

  if (!data) return <div style={{ padding: 12 }}>{err || 'Загрузка...'}</div>

  return (
    <div style={{ padding: 12 }}>
      <h2>{data.Program.Title}</h2>
      <Markdown text={data.Program.Description || ''} />
      {/* ✅ Кнопка "Опубликовать" (только admin) */}
      {ident.role === 'admin' && data?.Program?.ID && (
        <div style={{ marginTop: 10 }}>
          <AdminPublishProgram programId={data.Program.ID} />
        </div>
      )}
      {/* ВРЕМЕННО: дебаг, потом уберёшь
      <pre style={{ background: '#f6f6f6', padding: 8, overflow: 'auto' }}>
        {JSON.stringify(data, null, 2)}
      </pre> */}

      {ident.role === 'admin' && data?.Program?.ID && (
        <AdminCreateGroup programId={data.Program.ID} onDone={reload} />
      )}

      <h3>Открытые группы</h3>

      {(data.Groups ?? []).length === 0 ? (
        <div style={{ opacity: 0.75 }}>
          Пока нет групп. {ident.role === 'admin' ? 'Создай группу выше.' : ''}
        </div>
      ) : (
        <ul>
          {(data.Groups ?? []).map((g) => {
            const status = appByGroup.get(g.ID)
            const alreadyApplied = Boolean(status)

            return (
              <li key={g.ID} style={{ marginBottom: 14 }}>
                <div>
                  <b>{g.Title}</b> (мест: {g.Capacity}) {g.RequiresInterview ? '• интервью' : ''}
                </div>

                {alreadyApplied && (
                  <div style={{ opacity: 0.75 }}>
                    Заявка уже есть: <b>{status}</b>
                  </div>
                )}

                <div style={{ display: 'flex', gap: 8, flexWrap: 'wrap', marginTop: 6 }}>
                  <button disabled={alreadyApplied} onClick={() => apply(g.ID)}>
                    {alreadyApplied ? 'Заявка уже подана' : 'Подать заявку'}
                  </button>
                  <Link to={`/staff/groups/${g.ID}/applications`}>Заявки (staff)</Link>

                  <Link to={`/learn/group/${g.ID}`}>Перейти в обучение (если зачислен)</Link>
                </div>

                {/* ✅ Назначение преподавателей (admin/moderator видят список, назначает только admin) */}
                {(ident.role === 'admin' || ident.role === 'moderator') && (
                  <GroupTeacherAdminPanel groupId={g.ID} canAssign={ident.role === 'admin'} />
                )}

                {/* ✅ Редактирование группы (только admin) */}
                {ident.role === 'admin' && (
                  <div style={{ marginTop: 6 }}>
                    {!editing[g.ID] ? (
                      <button onClick={() => startEdit(g)}>Редактировать группу</button>
                    ) : (
                      <div style={{ marginTop: 8, padding: 10, border: '1px solid #ddd', borderRadius: 6 }}>
                        <div style={{ display: 'grid', gridTemplateColumns: '140px 1fr', gap: 8 }}>
                          <div>Название</div>
                          <input
                            value={draft[g.ID]?.title ?? ''}
                            onChange={(e) =>
                              setDraft((m) => ({ ...m, [g.ID]: { ...m[g.ID], title: e.target.value } }))
                            }
                          />

                          <div>Вместимость</div>
                          <input
                            type="number"
                            value={draft[g.ID]?.capacity ?? 0}
                            onChange={(e) =>
                              setDraft((m) => ({
                                ...m,
                                [g.ID]: { ...m[g.ID], capacity: Number(e.target.value) },
                              }))
                            }
                          />

                          <div>Открыта</div>
                          <label>
                            <input
                              type="checkbox"
                              checked={draft[g.ID]?.is_open ?? true}
                              onChange={(e) =>
                                setDraft((m) => ({
                                  ...m,
                                  [g.ID]: { ...m[g.ID], is_open: e.target.checked },
                                }))
                              }
                            />
                            {' '}да
                          </label>

                          <div>Интервью</div>
                          <label>
                            <input
                              type="checkbox"
                              checked={draft[g.ID]?.requires_interview ?? false}
                              onChange={(e) =>
                                setDraft((m) => ({
                                  ...m,
                                  [g.ID]: { ...m[g.ID], requires_interview: e.target.checked },
                                }))
                              }
                            />
                            {' '}требуется
                          </label>
                        </div>

                        <div style={{ display: 'flex', gap: 8, marginTop: 10 }}>
                          <button
                            disabled={saving[g.ID]}
                            onClick={async () => {
                              const d = draft[g.ID]
                              if (!d) return
                              setSaving((m) => ({ ...m, [g.ID]: true }))
                              try {
                                await api.updateGroup(g.ID, {
                                  title: d.title,
                                  capacity: d.capacity,
                                  is_open: d.is_open,
                                  requires_interview: d.requires_interview,
                                })
                                alert('Сохранено')
                                cancelEdit(g.ID)
                                await reload()
                              } catch (e: any) {
                                alert('Ошибка: ' + (e?.message || String(e)))
                              } finally {
                                setSaving((m) => ({ ...m, [g.ID]: false }))
                              }
                            }}
                          >
                            {saving[g.ID] ? 'Сохраняю...' : 'Сохранить'}
                          </button>

                          <button disabled={saving[g.ID]} onClick={() => cancelEdit(g.ID)}>
                            Отмена
                          </button>
                        </div>
                      </div>
                    )}
                  </div>
                )}
              </li>
            )
          })}
        </ul>
      )}
    </div>
  )
}
