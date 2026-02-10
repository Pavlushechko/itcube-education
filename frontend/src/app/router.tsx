// src/app/router.tsx

import { createBrowserRouter } from 'react-router-dom'
import { Layout } from './Layout'

import { CatalogPrograms } from '../pages/CatalogPrograms'
import { ProgramPage } from '../pages/ProgramPage'
import { MyApplications } from '../pages/MyApplications'
import { LearnGroup } from '../pages/LearnGroup'
import { TeacherGroups } from '../pages/TeacherGroups'
import { TeacherGroupStudents } from '../pages/TeacherGroupStudents'
import { TeacherGroupManage } from '../pages/TeacherGroupManage'
import { GroupApplicationsStaff } from '../pages/GroupApplicationsStaff.tsx'
import { ProgramApplications } from '../pages/ProgramApplications'
import { ApplicationsAll } from '../pages/ApplicationsAll'
import { TeacherInterview } from '../pages/TeacherInterview'


export const router = createBrowserRouter([
  {
    path: '/',
    element: <Layout />,
    children: [
      { index: true, element: <CatalogPrograms /> },
      { path: 'program/:id', element: <ProgramPage /> },
      { path: 'me/applications', element: <MyApplications /> },
      { path: 'learn/group/:groupId', element: <LearnGroup /> },
      { path: 'teacher/groups', element: <TeacherGroups /> },
      { path: 'teacher/groups/:groupId/students', element: <TeacherGroupStudents /> },
      { path: 'teacher/groups/:groupId/manage', element: <TeacherGroupManage /> },
      { path: '/staff/groups/:groupId/applications', element: <GroupApplicationsStaff /> },
      { path: 'program/:id/applications', element: <ProgramApplications /> },
      { path: 'applications', element: <ApplicationsAll /> },
      { path: 'teacher/applications/:appId/interview', element: <TeacherInterview /> },
    ],
  },
])
