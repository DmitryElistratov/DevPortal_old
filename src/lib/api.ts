const API_BASE_URL = (import.meta as any).env?.VITE_API_URL || 'http://localhost:8081/api';

class ApiClient {
  private token: string | null = null;

  constructor() {
    this.token = localStorage.getItem('auth_token');
  }

  setToken(token: string) {
    this.token = token;
    localStorage.setItem('auth_token', token);
  }

  clearToken() {
    this.token = null;
    localStorage.removeItem('auth_token');
  }

  private async request<T>(endpoint: string, options: RequestInit = {}): Promise<T> {
    const headers: Record<string, string> = {
      'Content-Type': 'application/json',
    };

    if (this.token) {
      headers['Authorization'] = `Bearer ${this.token}`;
    }

    const response = await fetch(`${API_BASE_URL}${endpoint}`, {
      ...options,
      headers,
    });

    if (!response.ok) {
      const error = await response.json().catch(() => ({ error: 'Unknown error' }));
      throw new Error(error.error || `HTTP ${response.status}`);
    }

    if (response.status === 204) {
      return {} as T;
    }

    return response.json();
  }

  // Auth
  async login(username: string, password: string) {
    const response = await this.request<{ token: string; user: any }>('/auth/login', {
      method: 'POST',
      body: JSON.stringify({ username, password }),
    });
    this.setToken(response.token);
    return response;
  }

  async register(data: { username: string; email: string; fullName: string; password: string; role?: string }) {
    const response = await this.request<{ token: string; user: any }>('/auth/register', {
      method: 'POST',
      body: JSON.stringify(data),
    });
    this.setToken(response.token);
    return response;
  }

  async validateToken(token: string): Promise<boolean> {
    try {
      this.token = token;
      await this.request('/me');
      return true;
    } catch {
      return false;
    }
  }

  async oauthLogin(data: { provider: string; externalId: string; username?: string; email?: string; fullName?: string }) {
    const response = await this.request<{ token: string; user: any }>('/auth/oauth', {
      method: 'POST',
      body: JSON.stringify(data),
    });
    this.setToken(response.token);
    return response;
  }

  async listUsers() {
    return this.request<any[]>('/users');
  }

  // Публичный (без авторизации) список включённых провайдеров авторизации —
  // используется страницей логина для кнопок «Войти через GitLab/GitHub/Keycloak».
  async listAuthProviders() {
    const response = await fetch(`${API_BASE_URL}/auth-providers`);
    if (!response.ok) return [];
    return response.json().catch(() => []);
  }

  async getMe() {
    return this.request<any>('/me');
  }

  // Services
  async listServices() {
    return this.request<any[]>('/services');
  }

  async getService(id: string) {
    return this.request<any>(`/services/${id}`);
  }

  async createService(data: any) {
    return this.request<any>('/services', {
      method: 'POST',
      body: JSON.stringify(data),
    });
  }

  async updateService(id: string, data: any) {
    return this.request<any>(`/services/${id}`, {
      method: 'PUT',
      body: JSON.stringify(data),
    });
  }

  async deleteService(id: string) {
    return this.request<void>(`/services/${id}`, { method: 'DELETE' });
  }

  // Releases
  async listReleases(serviceId?: string) {
    const query = serviceId ? `?serviceId=${serviceId}` : '';
    return this.request<any[]>(`/releases${query}`);
  }

  async getRelease(id: string) {
    return this.request<any>(`/releases/${id}`);
  }

  async createRelease(data: any) {
    return this.request<any>('/releases', {
      method: 'POST',
      body: JSON.stringify(data),
    });
  }

  async updateRelease(id: string, data: any) {
    return this.request<any>(`/releases/${id}`, {
      method: 'PUT',
      body: JSON.stringify(data),
    });
  }

  async deployRelease(id: string, environment: string) {
    return this.request<any>(`/releases/${id}/deploy`, {
      method: 'POST',
      body: JSON.stringify({ environment }),
    });
  }

  async rollbackRelease(id: string) {
    return this.request<any>(`/releases/${id}/rollback`, { method: 'POST' });
  }

  async approveRelease(id: string) {
    return this.request<any>(`/releases/${id}/approve`, { method: 'POST' });
  }

  // Audit
  async listAuditLogs(limit = 100, offset = 0) {
    return this.request<any[]>(`/audit?limit=${limit}&offset=${offset}`);
  }

  // Git Integration — реальные ветки/теги из GitLab через интеграцию портала.
  // project: числовой ID или URL-encoded path (group/project). Если не задан —
  // берётся из репозитория сервиса serviceId.
  async listGitBranches(projectId?: string, serviceId?: string) {
    const params = new URLSearchParams();
    if (projectId) params.set('project', projectId);
    if (serviceId) params.set('serviceId', serviceId);
    return this.request<any[]>(`/ci/branches?${params.toString()}`);
  }

  async listGitTags(projectId?: string, serviceId?: string) {
    const params = new URLSearchParams();
    if (projectId) params.set('project', projectId);
    if (serviceId) params.set('serviceId', serviceId);
    return this.request<any[]>(`/ci/tags?${params.toString()}`);
  }

  // Jira Integration
  async listJiraTasks(projectKey?: string) {
    const query = projectKey ? `?projectKey=${projectKey}` : '';
    return this.request<any[]>(`/jira/tasks${query}`);
  }

  async getJiraTask(taskKey: string) {
    return this.request<any>(`/jira/tasks/${taskKey}`);
  }

  async searchJiraTasks(jql: string) {
    return this.request<any[]>('/jira/tasks/search', {
      method: 'POST',
      body: JSON.stringify({ jql }),
    });
  }

  // Graylog Integration
  async searchGraylogLogs(query: string, timeRange: string = '1h', limit: number = 100) {
    return this.request<any[]>('/graylog/search', {
      method: 'POST',
      body: JSON.stringify({ query, timeRange, limit }),
    });
  }

  async getGraylogStreams() {
    return this.request<any[]>('/graylog/streams');
  }

  // Grafana Integration
  async getGrafanaDashboards() {
    return this.request<any[]>('/grafana/dashboards');
  }

  async getGrafanaMetrics(dashboardId: string, timeRange: string = '1h') {
    return this.request<any>(`/grafana/metrics/${dashboardId}?timeRange=${timeRange}`);
  }

  async getGrafanaDatasources() {
    return this.request<any[]>('/grafana/datasources');
  }

  // ArgoCD Integration
  async listArgoApplications() {
    return this.request<any[]>('/argocd/applications');
  }

  async syncArgoApplication(appName: string) {
    return this.request<any>(`/argocd/applications/${appName}/sync`, {
      method: 'POST',
    });
  }

  // Kubernetes API
  async listKubernetesPods(namespace?: string) {
    const query = namespace ? `?namespace=${namespace}` : '';
    return this.request<any[]>(`/kubernetes/pods${query}`);
  }

  async listKubernetesDeployments(namespace?: string) {
    const query = namespace ? `?namespace=${namespace}` : '';
    return this.request<any[]>(`/kubernetes/deployments${query}`);
  }

  async listKubernetesServices(namespace?: string) {
    const query = namespace ? `?namespace=${namespace}` : '';
    return this.request<any[]>(`/kubernetes/services${query}`);
  }

  async listKubernetesEvents(namespace?: string) {
    const query = namespace ? `?namespace=${namespace}` : '';
    return this.request<any[]>(`/kubernetes/events${query}`);
  }

  async getPodLogs(podName: string, namespace: string = 'default') {
    return this.request<string>(`/kubernetes/pods/${podName}/logs?namespace=${namespace}`);
  }

  async getKubernetesClusterInfo() {
    return this.request<any>('/kubernetes/cluster-info');
  }

  // Canary Deployments
  async createCanary(releaseId: string, percentage: number, argoApp: string) {
    return this.request<any>('/canary', {
      method: 'POST',
      body: JSON.stringify({ releaseId, percentage, argoApp }),
    });
  }

  async getCanary(id: string) {
    return this.request<any>(`/canary/${id}`);
  }

  async getCanaryByRelease(releaseId: string) {
    return this.request<any>(`/canary/release/${releaseId}`);
  }

  async getActiveCanary(serviceId: string) {
    return this.request<any>(`/canary/service/${serviceId}/active`);
  }

  async listCanaries(serviceId: string) {
    return this.request<any[]>(`/canary/service/${serviceId}/list`);
  }

  async updateCanaryPercentage(id: string, percentage: number) {
    return this.request<any>(`/canary/${id}/percentage`, {
      method: 'PUT',
      body: JSON.stringify({ percentage }),
    });
  }

  async promoteCanary(id: string) {
    return this.request<any>(`/canary/${id}/promote`, { method: 'POST' });
  }

  async rollbackCanary(id: string) {
    return this.request<any>(`/canary/${id}/rollback`, { method: 'POST' });
  }

  async listAllCanaries() {
    return this.request<any[]>('/canary');
  }

  async getActiveCanaryByRelease(releaseId: string) {
    return this.request<any>(`/canary/release/${releaseId}/active`);
  }

  async analyzeCanary(id: string) {
    return this.request<any>(`/canary/${id}/analyze`);
  }

  // Integrations
  async listIntegrations() {
    return this.request<any[]>('/integrations');
  }

  async getIntegration(id: string) {
    return this.request<any>(`/integrations/${id}`);
  }

  async createIntegration(data: any) {
    return this.request<any>('/integrations', {
      method: 'POST',
      body: JSON.stringify(data),
    });
  }

  async updateIntegration(id: string, data: any) {
    return this.request<any>(`/integrations/${id}`, {
      method: 'PUT',
      body: JSON.stringify(data),
    });
  }

  async deleteIntegration(id: string) {
    return this.request<void>(`/integrations/${id}`, { method: 'DELETE' });
  }

  async getIntegrationsByType(type: string) {
    return this.request<any[]>(`/integrations/type/${type}`);
  }

  async testIntegration(id: string) {
    return this.request<any>(`/integrations/${id}/test`, { method: 'POST' });
  }

  async syncIntegration(id: string) {
    return this.request<any>(`/integrations/${id}/sync`, { method: 'POST' });
  }

  // Active Versions (Multi-version deployment)
  async deployActiveVersion(data: { releaseId: string; environment: string; trafficPercentage: number }) {
    return this.request<any>('/active-versions', {
      method: 'POST',
      body: JSON.stringify(data),
    });
  }

  async getActiveVersions(serviceId: string, environment: string = 'production') {
    return this.request<any[]>(`/active-versions/service/${serviceId}?environment=${environment}`);
  }

  async updateVersionTraffic(versionId: string, trafficPercentage: number) {
    return this.request<any>(`/active-versions/${versionId}/traffic`, {
      method: 'PUT',
      body: JSON.stringify({ trafficPercentage }),
    });
  }

  async rollbackVersion(versionId: string) {
    return this.request<any>(`/active-versions/${versionId}/rollback`, { method: 'POST' });
  }

  async promoteVersion(versionId: string) {
    return this.request<any>(`/active-versions/${versionId}/promote`, { method: 'POST' });
  }

  // Vault Secrets Management
  async listVaultSecrets(path: string = 'secret') {
    return this.request<any[]>(`/vault/secrets?path=${path}`);
  }

  // Enriched listing for the tree view: returns entries with name/path/is_dir/
  // data/metadata in a single round trip (no N+1 GET per leaf).
  async listVaultEntries(path: string = 'secret') {
    return this.request<any[]>(`/vault/secrets?path=${encodeURIComponent(path)}&full=true`);
  }

  async getVaultSecret(path: string) {
    return this.request<any>(`/vault/secrets/${encodeURIComponent(path)}`);
  }

  async createVaultSecret(path: string, data: Record<string, any>) {
    return this.request<any>(`/vault/secrets/${encodeURIComponent(path)}`, {
      method: 'POST',
      body: JSON.stringify({ data }),
    });
  }

  async updateVaultSecret(path: string, data: Record<string, any>) {
    return this.request<any>(`/vault/secrets/${encodeURIComponent(path)}`, {
      method: 'PUT',
      body: JSON.stringify({ data }),
    });
  }

  async deleteVaultSecret(path: string) {
    return this.request<void>(`/vault/secrets/${encodeURIComponent(path)}`, { method: 'DELETE' });
  }

  async listVaultSecretVersions(path: string) {
    return this.request<any[]>(`/vault/secrets/${encodeURIComponent(path)}/versions`);
  }

  async restoreVaultSecretVersion(path: string, version: number) {
    return this.request<any>(`/vault/secrets/${encodeURIComponent(path)}/restore`, {
      method: 'POST',
      body: JSON.stringify({ version }),
    });
  }

  // ---------- TestOps ----------

  async listTestRuns(params?: { releaseId?: string; serviceId?: string }) {
    const qs = new URLSearchParams();
    if (params?.releaseId) qs.set('releaseId', params.releaseId);
    if (params?.serviceId) qs.set('serviceId', params.serviceId);
    const suffix = qs.toString() ? `?${qs.toString()}` : '';
    return this.request<any[]>(`/testruns${suffix}`);
  }

  async createTestRun(data: any) {
    return this.request<any>('/testruns', { method: 'POST', body: JSON.stringify(data) });
  }

  async deleteTestRun(id: string) {
    return this.request<any>(`/testruns/${id}`, { method: 'DELETE' });
  }

  async listChecklists() {
    return this.request<any[]>('/checklists');
  }

  async createChecklist(data: any) {
    return this.request<any>('/checklists', { method: 'POST', body: JSON.stringify(data) });
  }

  async updateChecklist(id: string, data: any) {
    return this.request<any>(`/checklists/${id}`, { method: 'PUT', body: JSON.stringify(data) });
  }

  async deleteChecklist(id: string) {
    return this.request<any>(`/checklists/${id}`, { method: 'DELETE' });
  }

  async createChecklistExecution(checklistId: string, releaseId: string) {
    return this.request<any>('/checklist-executions', {
      method: 'POST',
      body: JSON.stringify({ checklistId, releaseId }),
    });
  }

  async listChecklistExecutions(releaseId: string) {
    return this.request<any[]>(`/checklist-executions?releaseId=${encodeURIComponent(releaseId)}`);
  }

  async updateChecklistExecution(id: string, data: any) {
    return this.request<any>(`/checklist-executions/${id}`, { method: 'PUT', body: JSON.stringify(data) });
  }

  // ---------- CI/CD proxy ----------

  async listCIPipelines(project: string, ref?: string) {
    const qs = `project=${encodeURIComponent(project)}${ref ? `&ref=${encodeURIComponent(ref)}` : ''}`;
    return this.request<any[]>(`/ci/pipelines?${qs}`);
  }

  async triggerCIPipeline(project: string, ref: string, variables?: Record<string, string>) {
    return this.request<any>('/ci/pipelines', {
      method: 'POST',
      body: JSON.stringify({ project, ref, variables }),
    });
  }

  async listCIMergeRequests(project: string, state = 'opened') {
    return this.request<any[]>(`/ci/merge-requests?project=${encodeURIComponent(project)}&state=${state}`);
  }
}

export const api = new ApiClient();
