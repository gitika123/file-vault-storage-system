import React from 'react'
import { createRoot } from 'react-dom/client'
import App from './VaultWorkspace'
import './styles.css'
import './balkanid-brand.css'
import './modal.css'
import './filter.css'
createRoot(document.getElementById('root')!).render(<React.StrictMode><App /></React.StrictMode>)
