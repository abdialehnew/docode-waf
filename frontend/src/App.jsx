import { BrowserRouter as Router, Routes, Route } from 'react-router-dom'
import Layout from './components/Layout'
import Dashboard from './pages/Dashboard'
import VHosts from './pages/VHosts'
import IPGroups from './pages/IPGroups'
import TrafficLogs from './pages/TrafficLogs'
import Settings from './pages/Settings'

function App() {
  return (
    <Router>
      <Layout>
        <Routes>
          <Route path="/" element={<Dashboard />} />
          <Route path="/vhosts" element={<VHosts />} />
          <Route path="/ip-groups" element={<IPGroups />} />
          <Route path="/traffic" element={<TrafficLogs />} />
          <Route path="/settings" element={<Settings />} />
        </Routes>
      </Layout>
    </Router>
  )
}

export default App
