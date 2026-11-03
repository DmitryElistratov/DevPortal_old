import { HashRouter, Routes, Route, Navigate } from 'react-router-dom';
import { AuthProvider, useAuth } from './contexts/AuthContext';
import { SettingsProvider } from './contexts/SettingsContext';
import { PluginProvider } from './contexts/PluginContext';
import Layout from './components/Layout';
import { ErrorBoundary } from './components/ErrorBoundary';
import { ReactNode, Suspense, lazy } from 'react';

// Lazy load pages
const HomePage = lazy(() => import('./pages/HomePage'));
const CatalogPage = lazy(() => import('./pages/CatalogPage'));
const ComponentPage = lazy(() => import('./pages/ComponentPage'));
const ReleasePage = lazy(() => import('./pages/ReleasePage'));
const ReleasesPage = lazy(() => import('./pages/ReleasesPage'));
const AuditPage = lazy(() => import('./pages/AuditPage'));
const TTMPage = lazy(() => import('./pages/TTMPage'));
const ChecklistsPage = lazy(() => import('./pages/ChecklistsPage'));
const CreateReleasePage = lazy(() => import('./pages/CreateReleasePage'));
const TemplatesPage = lazy(() => import('./pages/TemplatesPage'));
const DocsPage = lazy(() => import('./pages/DocsPage'));
const SettingsPage = lazy(() => import('./pages/SettingsPage'));
const LoginPage = lazy(() => import('./pages/LoginPage'));
const DeploymentsPage = lazy(() => import('./pages/DeploymentsPage'));
const PluginsPage = lazy(() => import('./pages/PluginsPage'));
const KubeAPIPage = lazy(() => import('./pages/KubeAPIPage'));
const ObservabilityPage = lazy(() => import('./pages/ObservabilityPage'));
const CanaryDeploymentPage = lazy(() => import('./pages/CanaryDeploymentPage'));
const ActiveVersionsPage = lazy(() => import('./pages/ActiveVersionsPage'));
const VaultSecretsPage = lazy(() => import('./pages/VaultSecretsPage'));

function ProtectedRoute({ children }: { children: ReactNode }) {
  const { isAuthenticated, loading } = useAuth();

  if (loading) {
    return (
      <div className="min-h-screen flex items-center justify-center bg-gray-50">
        <div className="text-gray-500">Загрузка...</div>
      </div>
    );
  }

  if (!isAuthenticated) {
    return <Navigate to="/login" replace />;
  }

  return <>{children}</>;
}

const LoadingFallback = () => (
  <div className="min-h-screen flex items-center justify-center bg-gray-50">
    <div className="text-gray-500">Загрузка...</div>
  </div>
);

function AppRoutes() {
  return (
    <Suspense fallback={<LoadingFallback />}>
      <Routes>
        <Route path="/login" element={<LoginPage />} />
        <Route path="/" element={<ProtectedRoute><Layout><HomePage /></Layout></ProtectedRoute>} />
        <Route path="/catalog" element={<ProtectedRoute><Layout><CatalogPage /></Layout></ProtectedRoute>} />
        <Route path="/service/:serviceId" element={<ProtectedRoute><Layout><ComponentPage /></Layout></ProtectedRoute>} />
        <Route path="/service/:serviceId/create-release" element={<ProtectedRoute><Layout><CreateReleasePage /></Layout></ProtectedRoute>} />
        <Route path="/service/:serviceId/observability" element={<ProtectedRoute><Layout><ObservabilityPage /></Layout></ProtectedRoute>} />
        <Route path="/service/:serviceId/active-versions" element={<ProtectedRoute><Layout><ActiveVersionsPage /></Layout></ProtectedRoute>} />
        <Route path="/release/:releaseId" element={<ProtectedRoute><Layout><ReleasePage /></Layout></ProtectedRoute>} />
        <Route path="/release/:releaseId/canary" element={<ProtectedRoute><Layout><CanaryDeploymentPage /></Layout></ProtectedRoute>} />
        <Route path="/canary" element={<ProtectedRoute><Layout><CanaryDeploymentPage /></Layout></ProtectedRoute>} />
        <Route path="/releases" element={<ProtectedRoute><Layout><ReleasesPage /></Layout></ProtectedRoute>} />
        <Route path="/audit" element={<ProtectedRoute><Layout><AuditPage /></Layout></ProtectedRoute>} />
        <Route path="/ttm" element={<ProtectedRoute><Layout><TTMPage /></Layout></ProtectedRoute>} />
        <Route path="/checklists" element={<ProtectedRoute><Layout><ChecklistsPage /></Layout></ProtectedRoute>} />
        <Route path="/templates" element={<ProtectedRoute><Layout><TemplatesPage /></Layout></ProtectedRoute>} />
        <Route path="/docs" element={<ProtectedRoute><Layout><DocsPage /></Layout></ProtectedRoute>} />
        <Route path="/settings" element={<ProtectedRoute><Layout><SettingsPage /></Layout></ProtectedRoute>} />
        <Route path="/deployments" element={<ProtectedRoute><Layout><DeploymentsPage /></Layout></ProtectedRoute>} />
        <Route path="/plugins" element={<ProtectedRoute><Layout><PluginsPage /></Layout></ProtectedRoute>} />
        <Route path="/kube-api" element={<ProtectedRoute><Layout><KubeAPIPage /></Layout></ProtectedRoute>} />
        <Route path="/vault/secrets" element={<ProtectedRoute><Layout><VaultSecretsPage /></Layout></ProtectedRoute>} />
      </Routes>
    </Suspense>
  );
}

function App() {
  return (
    <ErrorBoundary>
      <HashRouter>
        <AuthProvider>
          <SettingsProvider>
            <PluginProvider>
              <AppRoutes />
            </PluginProvider>
          </SettingsProvider>
        </AuthProvider>
      </HashRouter>
    </ErrorBoundary>
  );
}

export default App;
