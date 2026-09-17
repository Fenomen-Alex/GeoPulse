/* @refresh reload */
import { render } from 'solid-js/web'
import '@fontsource-variable/inter'
import '@fontsource-variable/jetbrains-mono'
import 'maplibre-gl/dist/maplibre-gl.css'
import './index.css'
import App from './App.tsx'

const root = document.getElementById('root')

render(() => <App />, root!)
