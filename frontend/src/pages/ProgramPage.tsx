
// src/pages/ProgramPage.tsx
import { useEffect, useMemo, useState } from 'react'
import { Link, useParams } from 'react-router-dom'
import { api } from '../lib/api'
import type { ProgramWithGroups } from '../lib/types'
import { getIdentity } from '../lib/auth'
import { Markdown } from '../ui/Markdown'
import { AdminPublishProgram } from '../ui/AdminPublishProgram'
import { AdminCreateGroup } from '../ui/AdminCreateGroup'
import { GroupTeacherAdminPanel } from '../ui/GroupTeacherAdminPanel'

type AppLite = {
  GroupID: string
  Status: string
}

type PrivateProgramView = {
  program: any
  cohorts: any[]
  groups: any[]
}

export function ProgramPage() {
  const { id } = useParams()
  const ident = getIdentity()

  const isStaff = ident.role === 'admin' || ident.role === 'moderator'
  const isAdmin = ident.role === 'admin'

  const [data, setData] = useState<ProgramWithGroups | null>(null)
  const [err, setErr] = useState('')

  // teacherMode = это когда role=user, но доступ к private /programs/{id} есть
  // (значит назначен преподавателем хотя бы в одну группу этого курса)
  const [teacherMode, setTeacherMode] = useState(false)

  const [myApps, setMyApps] = useState<AppLite[]>([])

  // admin edit
  const [editTitle, setEditTitle] = useState('')
  const [editDesc, setEditDesc] = useState('')
  const [savingProgram, setSavingProgram] = useState(false)

  const appByGroup = useMemo(() => {
    const m = new Map<string, string>()
    for (const a of myApps) m.set(a.GroupID, a.Status)
    return m
  }, [myApps])

  async function reload() {
    if (!id) return
    setErr('')

    try {
      // 1) Грузим данные программы
      if (isStaff) {
        const pg = await api.getProgramAdmin(id)
        setData(pg)
        setTeacherMode(false)
      } else {
        // для user всегда public (никаких 403)
        const pg = await api.getProgram(id)
        setData(pg)

        // 2) Проверяем назначение "преподаватель по программе" отдельной ручкой (200 ok:false)
        try {
          const res = await api.teacherProgramAccess(id)
          setTeacherMode(Boolean(res?.ok))
        } catch {
          // если вдруг ручка недоступна — просто считаем не учитель
          setTeacherMode(false)
        }
      }

      // 3) Мои заявки
      if (ident.role === 'user') {
        const apps = await api.listMyApplications()
        const lite: AppLite[] = (apps ?? []).map((a: any) => ({
          GroupID: a.GroupID,
          Status: a.Status,
        }))
        setMyApps(lite)
      } else {
        setMyApps([])
      }
    } catch (e: any) {
      setErr(String(e.message || e))
    }
  }


  useEffect(() => {
    setData(null)
    setTeacherMode(false)
    reload()
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [id, isStaff])

  useEffect(() => {
    if (!data) return
    setEditTitle(data.Program.Title || '')
    setEditDesc(data.Program.Description || '')
  }, [data])

  async function apply(groupId: string) {
    if (appByGroup.has(groupId)) return
    try {
      const res = await api.createApplication(groupId, 'хочу учиться')
      alert('Заявка создана: ' + res.id)
      await reload()
    } catch (e: any) {
      alert('Ошибка: ' + (e?.message || String(e)))
    }
  }

  async function saveProgram() {
    if (!id) return
    setSavingProgram(true)
    try {
      await api.updateProgram(id, { title: editTitle, description: editDesc })
      alert('Сохранено')
      await reload()
    } catch (e: any) {
      alert('Ошибка: ' + (e?.message || String(e)))
    } finally {
      setSavingProgram(false)
    }
  }

  if (!data) return <div>{err || 'Загрузка...'}</div>

  const canSeeProgramAppsButton = teacherMode || isStaff

  return (
    <div>
      <h2>{data.Program.Title}</h2>

      <Markdown text={data.Program.Description || ''} />

      {canSeeProgramAppsButton && id && (
        <div>
          <Link to={`/program/${id}/applications`}>Заявки на курс</Link>
        </div>
      )}

      {isAdmin && id && (
        <div>
          <h3>Управление программой (admin)</h3>

          <div>
            <div>Название</div>
            <input value={editTitle} onChange={(e) => setEditTitle(e.target.value)} />
          </div>

          <div>
            <div>Описание</div>
            <textarea value={editDesc} onChange={(e) => setEditDesc(e.target.value)} />
          </div>

          <div>
            <button disabled={savingProgram} onClick={saveProgram}>
              {savingProgram ? 'Сохраняю...' : 'Сохранить'}
            </button>
          </div>

          <div>
            <AdminPublishProgram programId={id} />
          </div>

          <div>
            <AdminCreateGroup programId={id} onDone={reload} />
          </div>
        </div>
      )}

      <h3>Группы</h3>

      {(data.Groups ?? []).length === 0 ? (
        <div>Пока нет групп.</div>
      ) : (
        <ul>
          {(data.Groups ?? []).map((g: any) => {
            const status = appByGroup.get(g.ID)
            const alreadyApplied = Boolean(status)

            return (
              <li key={g.ID}>
                <div>
                  <b>{g.Title}</b>
                </div>
                <div>
                  Мест: {g.Capacity} {g.RequiresInterview ? '• интервью' : ''}
                  {isStaff ? (g.IsOpen ? ' • открыта' : ' • закрыта') : null}
                </div>

                {ident.role === 'user' && !teacherMode && (
                  <div>
                    {alreadyApplied ? (
                      <div>Заявка уже есть: {status}</div>
                    ) : (
                      <button onClick={() => apply(g.ID)}>Подать заявку</button>
                    )}

                    <div>
                      <Link to={`/learn/group/${g.ID}`}>Перейти в обучение (если зачислен)</Link>
                    </div>
                  </div>
                )}

                {(ident.role === 'admin' || ident.role === 'moderator') && (
                  <div>
                    <GroupTeacherAdminPanel groupId={g.ID} canAssign={ident.role === 'admin'} />
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
