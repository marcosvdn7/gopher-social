import { StrictMode } from 'react'
import { createRoot } from 'react-dom/client'
import './index.css'
import { ConfirmationPage } from './ConfirmationPage.tsx'
import { createBrowserRouter, RouterProvider } from 'react-router-dom'
import { App } from './App.tsx'
import { CreatePostForm } from './CreatePostForms.tsx'


const router = createBrowserRouter([
  {
    path: "/",
    element: <App />
  },
  {
    path: "/confirm/:token",
    element: <ConfirmationPage />
  },
  {
    path: "/createpost",
    element: <CreatePostForm></CreatePostForm>
  }
])

createRoot(document.getElementById('root')!).render(
  <StrictMode>
    <RouterProvider router={router} />
  </StrictMode>,
)
