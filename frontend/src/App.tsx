// Copyright (c) 2025 kk
//
// This software is released under the MIT License.
// https://opensource.org/licenses/MIT

import { BrowserRouter, Routes, Route, Navigate } from 'react-router-dom';
import { AuthProvider } from './contexts/AuthContext';
import ErrorBoundary from './components/ErrorBoundary';
import ProtectedRoute from './components/ProtectedRoute';
import Layout from './components/Layout';
import LoginPage from './pages/LoginPage';
import RulesPage from './pages/RulesPage';
import AlertsPage from './pages/AlertsPage';
import DashboardPage from './pages/DashboardPage';
import RuleEditPage from './pages/RuleEditPage';
import ESConfigPage from './pages/ESConfigPage';
import LarkConfigPage from './pages/LarkConfigPage';
import CleanupConfigPage from './pages/CleanupConfigPage';
import UsersPage from './pages/UsersPage';
import SsoCallbackPage from './pages/SsoCallbackPage';
import SsoConfigPage from './pages/SsoConfigPage';
import AdminRoute from './components/AdminRoute';

function App() {
  return (
    <ErrorBoundary>
      <AuthProvider>
        <BrowserRouter>
          <Routes>
            <Route path="/login" element={<LoginPage />} />
            <Route path="/sso-callback" element={<SsoCallbackPage />} />
            <Route
              path="/*"
              element={
                <ProtectedRoute>
                  <Layout>
                    <Routes>
                      <Route path="/" element={<DashboardPage />} />
                      <Route path="/rules" element={<RulesPage />} />
                      <Route path="/rules/new" element={<RuleEditPage />} />
                      <Route path="/rules/:id/edit" element={<RuleEditPage />} />
                      <Route path="/alerts" element={<AlertsPage />} />
                      <Route path="/es-configs" element={<ESConfigPage />} />
                      <Route path="/lark-configs" element={<LarkConfigPage />} />
                      <Route path="/cleanup-config" element={<CleanupConfigPage />} />
                      <Route
                        path="/users"
                        element={
                          <AdminRoute>
                            <UsersPage />
                          </AdminRoute>
                        }
                      />
                      <Route
                        path="/sso-config"
                        element={
                          <AdminRoute>
                            <SsoConfigPage />
                          </AdminRoute>
                        }
                      />
                      <Route path="*" element={<Navigate to="/" replace />} />
                    </Routes>
                  </Layout>
                </ProtectedRoute>
              }
            />
          </Routes>
        </BrowserRouter>
      </AuthProvider>
    </ErrorBoundary>
  );
}

export default App;
