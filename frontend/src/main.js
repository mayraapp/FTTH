import { mountOperationsApp } from './App.js?v=28'
import { operationPage } from './pages/operations.js'

document.documentElement.dataset.page = operationPage.id
mountOperationsApp('#app')
