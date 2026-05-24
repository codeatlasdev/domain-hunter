import { StrictMode } from 'react'
import { createRoot } from 'react-dom/client'
import { RouterProvider, createRouter, createRoute, createRootRoute } from '@tanstack/react-router'
import { Toaster } from 'sonner'
import { Layout } from './components/layout'
import { Home } from './pages/home'
import { TLDs } from './pages/tlds'
import { Docs } from './pages/docs'
import './index.css'

const rootRoute = createRootRoute({ component: Layout })
const indexRoute = createRoute({ getParentRoute: () => rootRoute, path: '/', component: Home })
const tldsRoute = createRoute({ getParentRoute: () => rootRoute, path: '/tlds', component: TLDs })
const docsRoute = createRoute({ getParentRoute: () => rootRoute, path: '/docs', component: Docs })

const routeTree = rootRoute.addChildren([indexRoute, tldsRoute, docsRoute])
const router = createRouter({ routeTree })

declare module '@tanstack/react-router' {
  interface Register { router: typeof router }
}

createRoot(document.getElementById('root')!).render(
  <StrictMode>
    <RouterProvider router={router} />
    <Toaster position="bottom-right" />
  </StrictMode>,
)
