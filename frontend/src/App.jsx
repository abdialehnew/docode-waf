import { BrowserRouter as Router, Routes, Route, Navigate } from 'react-router-dom'
import { AuthProvider } from './context/AuthContext'
import ProtectedRoute from './components/ProtectedRoute'
import Layout from './components/Layout'
import Login from './pages/Login'
import Register from './pages/Register'
import ForgotPassword from './pages/ForgotPassword'
import Dashboard from './pages/Dashboard'
import VHosts from './pages/VHosts'
import VHostForm from './pages/VHostForm'
import IPGroups from './pages/IPGroups'
import TrafficLogs from './pages/TrafficLogs'
import Certificates from './pages/Certificates'
import Settings from './pages/Settings'
import Monitoring from './pages/Monitoring'
import Bans from './pages/Bans.jsx'
import NotificationSettings from './pages/NotificationSettings'
import Rules from './pages/Rules'
import Reports from './pages/Reports.jsx'

function App() {
  return (
    <Router>
      <AuthProvider>
        <Routes>
          {/* Public routes */}
          <Route path="/login" element={<Login />} />
          <Route path="/register" element={<Register />} />
          <Route path="/forgot-password" element={<ForgotPassword />} />

          {/* Protected routes */}
          <Route path="/" element={
            <ProtectedRoute>
              <Layout>
                <Dashboard />
              </Layout>
            </ProtectedRoute>
          } />
          <Route path="/vhosts" element={
            <ProtectedRoute>
              <Layout>
                <VHosts />
              </Layout>
            </ProtectedRoute>
          } />
          <Route path="/vhosts/new" element={
            <ProtectedRoute>
              <Layout>
                <VHostForm />
              </Layout>
            </ProtectedRoute>
          } />
          <Route path="/vhosts/:id/edit" element={
            <ProtectedRoute>
              <Layout>
                <VHostForm />
              </Layout>
            </ProtectedRoute>
          } />
          <Route path="/ip-groups" element={
            <ProtectedRoute>
              <Layout>
                <IPGroups />
              </Layout>
            </ProtectedRoute>
          } />
          <Route path="/bans" element={
            <ProtectedRoute>
              <Layout>
                <Bans />
              </Layout>
            </ProtectedRoute>
          } />
          <Route path="/traffic" element={
            <ProtectedRoute>
              <Layout>
                <TrafficLogs />
              </Layout>
            </ProtectedRoute>
          } />
          <Route path="/certificates" element={
            <ProtectedRoute>
              <Layout>
                <Certificates />
              </Layout>
            </ProtectedRoute>
          } />
          <Route path="/monitoring" element={
            <ProtectedRoute>
              <Layout>
                <Monitoring />
              </Layout>
            </ProtectedRoute>
          } />
          <Route path="/settings" element={
            <ProtectedRoute>
              <Layout>
                <Settings />
              </Layout>
            </ProtectedRoute>
          } />
          <Route path="/notifications" element={
            <ProtectedRoute>
              <Layout>
                <NotificationSettings />
              </Layout>
            </ProtectedRoute>
          } />
          <Route path="/rules" element={
            <ProtectedRoute>
              <Layout>
                <Rules />
              </Layout>
            </ProtectedRoute>
          } />
          <Route path="/reports" element={
            <ProtectedRoute>
              <Layout>
                <Reports />
              </Layout>
            </ProtectedRoute>
          } />

          {/* Fallback route */}
          <Route path="*" element={<Navigate to="/" replace />} />
        </Routes>
      </AuthProvider>
    </Router>
  )
}

export default App
