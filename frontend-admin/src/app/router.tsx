// src/app/router.tsx
import { createBrowserRouter, Navigate } from 'react-router-dom'
import { Layout } from './Layout'

import { CatalogPrograms } from '../pages/CatalogPrograms'
import { ProgramPage } from '../pages/ProgramPage'

import { ApplicationsAll } from '../pages/ApplicationsAll'
import { GroupApplicationsStaff } from '../pages/GroupApplicationsStaff'
import { ProgramApplications } from '../pages/ProgramApplications'

export const router = createBrowserRouter([
  {
    path: '/',
    element: <Layout />,
    children: [
      { index: true, element: <Navigate to="/catalog" replace /> },
      { path: 'catalog', element: <CatalogPrograms /> },

      { path: 'program/:id', element: <ProgramPage /> },

      { path: 'program/:id/applications', element: <ProgramApplications /> },

      { path: 'applications', element: <ApplicationsAll /> },
      { path: 'staff/groups/:groupId/applications', element: <GroupApplicationsStaff /> },
    ],
  },
])
