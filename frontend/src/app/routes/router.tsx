import { createBrowserRouter, Navigate } from 'react-router-dom'

import { AppLayout } from '../../components/layout/AppLayout'
import { DashboardPage } from '../../pages/DashboardPage'
import { LoginPage } from '../../pages/LoginPage'
import { NotFoundPage } from '../../pages/NotFoundPage'
import { RequestCreatePage } from '../../pages/RequestCreatePage'
import { RequestDetailPage } from '../../pages/RequestDetailPage'
import { RequestEditPage } from '../../pages/RequestEditPage'
import { RequestsPage } from '../../pages/RequestsPage'
import { ProtectedRoute } from './ProtectedRoute'
import { PublicOnlyRoute } from './PublicOnlyRoute'

export const router = createBrowserRouter([
  {
    element: <PublicOnlyRoute />,
    children: [{ path: '/login', element: <LoginPage /> }],
  },
  {
    element: <ProtectedRoute />,
    children: [
      {
        element: <AppLayout />,
        children: [
          { index: true, element: <Navigate to="/dashboard" replace /> },
          { path: '/dashboard', element: <DashboardPage /> },
          { path: '/solicitacoes', element: <RequestsPage /> },
          { path: '/solicitacoes/nova', element: <RequestCreatePage /> },
          { path: '/solicitacoes/:id', element: <RequestDetailPage /> },
          { path: '/solicitacoes/:id/editar', element: <RequestEditPage /> },
        ],
      },
    ],
  },
  { path: '*', element: <NotFoundPage /> },
])
