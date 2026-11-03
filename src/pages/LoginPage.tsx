import { useState, useEffect } from 'react';
import { useNavigate } from 'react-router-dom';
import { useAuth } from '../contexts/AuthContext';
import { api } from '../lib/api';
import { Lock, User, AlertCircle, Loader2 } from 'lucide-react';

type OAuthProvider = 'gitlab' | 'github' | 'keycloak';

interface ProviderButtonConfig {
  provider: OAuthProvider;
  label: string;
  authUrl: string;
  tokenUrl: string;
  profileUrl: string;
  scope: string;
  clientId: string;
}

const PROVIDER_LABELS: Record<OAuthProvider, string> = {
  gitlab: 'GitLab',
  github: 'GitHub',
  keycloak: 'Keycloak',
};

// Извлекает access_token / error из query- или hash-фрагмента URL возврата.
function extractOAuthParams(): { code?: string; access_token?: string; error?: string } {
  const result: { code?: string; access_token?: string; error?: string } = {};
  const parse = (source: string) => {
    new URLSearchParams(source).forEach((value, key) => {
      if (key === 'code' || key === 'access_token' || key === 'error') {
        result[key] = value;
      }
    });
  };
  parse(window.location.search.replace(/^\?/, ''));
  if (!result.code && !result.access_token && !result.error) {
    parse(window.location.hash.replace(/^#/, ''));
  }
  return result;
}

async function fetchProfile(provider: OAuthProvider, baseUrl: string, accessToken: string): Promise<any> {
  const headers = new Headers({ Authorization: `Bearer ${accessToken}` });
  if (provider === 'github') headers.set('Accept', 'application/vnd.github+json');
  const response = await fetch(`${baseUrl.replace(/\/+$/, '')}/api/v4/user`, { headers }).catch(() => null);
  // GitHub использует другой endpoint — пробуем его, если GitLab-стиль не сработал.
  if (provider === 'github') {
    const gh = await fetch('https://api.github.com/user', { headers }).catch(() => null);
    if (gh?.ok) {
      const data = await gh.json();
      return { id: String(data.id), username: data.login, name: data.name || data.login, email: data.email || '' };
    }
  } else if (response?.ok) {
    const data = await response.json();
    return { id: String(data.id), username: data.username, name: data.name || data.username, email: data.public_email || data.email || '' };
  }
  if (provider === 'keycloak' && baseUrl) {
    const wellKnown = `${baseUrl.replace(/\/+$/, '')}/protocol/openid-connect/userinfo`;
    const ui = await fetch(wellKnown, { headers }).catch(() => null);
    if (ui?.ok) {
      const data = await ui.json();
      return {
        id: data.sub,
        username: data.preferred_username || data.sub,
        name: data.name || data.preferred_username || '',
        email: data.email || '',
      };
    }
  }
  throw new Error('Не удалось получить профиль пользователя у провайдера');
}

export default function LoginPage() {
  const { login } = useAuth();
  const navigate = useNavigate();
  const [username, setUsername] = useState('');
  const [password, setPassword] = useState('');
  const [error, setError] = useState('');
  const [loading, setLoading] = useState(false);
  const [providers, setProviders] = useState<ProviderButtonConfig[]>([]);
  const [oauthLoading, setOauthLoading] = useState<string | null>(null);

  // Обработка возврата из OAuth-провайдера (?code=... или #access_token=...)
  useEffect(() => {
    const params = extractOAuthParams();
    if (!(params.code || params.access_token)) return;
    const savedRaw = sessionStorage.getItem('oauth_flow');
    sessionStorage.removeItem('oauth_flow');
    if (!savedRaw) return;
    const saved = JSON.parse(savedRaw) as ProviderButtonConfig;
    if (params.error) {
      setError(`Ошибка авторизации через ${PROVIDER_LABELS[saved.provider]}: ${params.error}`);
      window.history.replaceState({}, '', '/login');
      return;
    }
    let cancelled = false;
    (async () => {
      try {
        setOauthLoading(saved.provider);
        let accessToken = params.access_token;
        if (!accessToken && params.code) {
          // Конвертация authorization code в токен требует client_secret,
          // который хранится только на бэкенде.
          setError(
            'Получен authorization code, но обмен на токен должен выполнять бэкенд (endpoint ещё не реализован).'
          );
          return;
        }
        if (!accessToken) throw new Error('В ответе провайдера отсутствует access_token');
        const profile = await fetchProfile(saved.provider, saved.authUrl, accessToken);
        await api.oauthLogin({
          provider: saved.provider,
          externalId: String(profile.id),
          username: profile.username,
          email: profile.email,
          fullName: profile.name,
        });
        const user = await api.getMe();
        localStorage.setItem('release_portal_user', JSON.stringify(user));
        window.history.replaceState({}, '', '/login');
        navigate('/');
        window.location.reload();
      } catch (e: any) {
        setError(`Не удалось завершить вход через ${PROVIDER_LABELS[saved.provider]}: ${e?.message || e}`);
        window.history.replaceState({}, '', '/login');
      } finally {
        setOauthLoading(null);
      }
    })();
    return () => { cancelled = true; };
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, []);

  // Список включённых провайдеров авторизации берём из настроек портала.
  useEffect(() => {
    let cancelled = false;
    (async () => {
      try {
        const list: any[] = await api.listIntegrations().catch(() => []);
        if (cancelled) return;
        const configs: ProviderButtonConfig[] = [];
        for (const item of list || []) {
          if (item.type !== 'auth' || !item.enabled) continue;
          const rawProvider: string = item.config?.provider || 'keycloak';
          if (rawProvider === 'ldap') continue; // LDAP логины обрабатывает локальная форма
          if (rawProvider !== 'gitlab' && rawProvider !== 'github' && rawProvider !== 'keycloak') continue;
          const provider: OAuthProvider = rawProvider;
          const baseUrl = (item.url || '').replace(/\/+$/, '');
          let authUrl = '';
          let tokenUrl = '';
          if (provider === 'gitlab') {
            authUrl = `${baseUrl}/oauth/authorize`;
            tokenUrl = `${baseUrl}/oauth/token`;
          } else if (provider === 'github') {
            authUrl = `${baseUrl || 'https://github.com'}/login/oauth/authorize`;
            tokenUrl = `${baseUrl || 'https://github.com'}/login/oauth/access_token`;
          } else if (provider === 'keycloak') {
            const realm = item.config?.realm || '';
            const root = baseUrl.includes('/realms/') ? baseUrl : `${baseUrl}/realms/${realm}`;
            authUrl = `${root}/protocol/openid-connect/auth`;
            tokenUrl = `${root}/protocol/openid-connect/token`;
          }
          if (!authUrl || !item.config?.clientId) continue;
          configs.push({
            provider,
            label: item.name || PROVIDER_LABELS[provider],
            authUrl,
            tokenUrl,
            profileUrl: baseUrl,
            scope: provider === 'gitlab' ? 'read_user openid' : provider === 'github' ? 'read:user user:email' : 'openid profile email',
            clientId: item.config.clientId,
          });
        }
        if (!cancelled) setProviders(configs);
      } catch {
        /* список провайдеров не критичен для локального входа */
      }
    })();
    return () => { cancelled = true; };
  }, []);

  const handleOAuthLogin = (config: ProviderButtonConfig) => {
    setError('');
    sessionStorage.setItem('oauth_flow', JSON.stringify(config));
    const redirectUri = `${window.location.origin}${window.location.pathname}`;
    const state = Math.random().toString(36).slice(2);
    const url =
      `${config.authUrl}?client_id=${encodeURIComponent(config.clientId)}` +
      `&redirect_uri=${encodeURIComponent(redirectUri)}` +
      `&response_type=token&scope=${encodeURIComponent(config.scope)}&state=${state}`;
    window.location.href = url;
  };

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    setError('');
    setLoading(true);

    const success = await login(username, password);
    
    if (success) {
      navigate('/');
    } else {
      setError('Неверное имя пользователя или пароль');
      setLoading(false);
    }
  };

  return (
    <div className="min-h-screen flex items-center justify-center bg-black">
      <div className="w-full max-w-md">
        <div className="bg-white rounded-lg shadow-2xl p-8">
          <div className="text-center mb-8">
            <div className="w-16 h-16 bg-orange-500 rounded-full flex items-center justify-center mx-auto mb-4">
              <span className="text-white font-bold text-2xl">R</span>
            </div>
            <h1 className="text-2xl font-semibold text-gray-900">Release Portal</h1>
            <p className="text-gray-500 text-sm mt-1">Release Management System</p>
          </div>

          <form onSubmit={handleSubmit} className="space-y-6">
            {error && (
              <div className="bg-red-50 border border-red-200 rounded-lg p-3 flex items-start gap-2">
                <AlertCircle className="w-5 h-5 text-red-500 flex-shrink-0 mt-0.5" />
                <span className="text-sm text-red-700">{error}</span>
              </div>
            )}

            <div>
              <label className="block text-sm font-medium text-gray-700 mb-2">
                Имя пользователя
              </label>
              <div className="relative">
                <User className="absolute left-3 top-1/2 -translate-y-1/2 w-5 h-5 text-gray-400" />
                <input
                  type="text"
                  value={username}
                  onChange={(e) => setUsername(e.target.value)}
                  className="w-full pl-10 pr-4 py-2 border border-gray-300 rounded-lg focus:outline-none focus:ring-2 focus:ring-orange-500 focus:border-transparent"
                  placeholder="admin"
                  required
                />
              </div>
            </div>

            <div>
              <label className="block text-sm font-medium text-gray-700 mb-2">
                Пароль
              </label>
              <div className="relative">
                <Lock className="absolute left-3 top-1/2 -translate-y-1/2 w-5 h-5 text-gray-400" />
                <input
                  type="password"
                  value={password}
                  onChange={(e) => setPassword(e.target.value)}
                  className="w-full pl-10 pr-4 py-2 border border-gray-300 rounded-lg focus:outline-none focus:ring-2 focus:ring-orange-500 focus:border-transparent"
                  placeholder="••••••••"
                  required
                />
              </div>
            </div>

            <button
              type="submit"
              disabled={loading}
              className="w-full bg-orange-500 hover:bg-orange-600 disabled:bg-gray-400 text-white font-medium py-2 px-4 rounded-lg transition-colors"
            >
              {loading ? 'Вход...' : 'Войти'}
            </button>
          </form>

          {providers.length > 0 && (
            <div className="mt-6">
              <div className="relative flex items-center justify-center mb-4">
                <div className="absolute w-full border-t border-gray-200" />
                <span className="relative bg-white px-3 text-xs text-gray-500">или войдите через</span>
              </div>
              <div className="space-y-2">
                {providers.map(config => (
                  <button
                    key={`${config.provider}-${config.label}`}
                    type="button"
                    onClick={() => handleOAuthLogin(config)}
                    disabled={!!oauthLoading}
                    className="w-full flex items-center justify-center gap-2 border border-gray-300 hover:bg-gray-50 disabled:bg-gray-100 text-gray-700 font-medium py-2 px-4 rounded-lg transition-colors text-sm"
                  >
                    {oauthLoading === config.provider ? (
                      <Loader2 className="w-4 h-4 animate-spin" />
                    ) : (
                      <span className={`w-4 h-4 rounded-full ${
                        config.provider === 'gitlab' ? 'bg-orange-600' :
                        config.provider === 'github' ? 'bg-gray-900' : 'bg-blue-600'
                      }`} />
                    )}
                    Войти через {config.label}
                  </button>
                ))}
              </div>
            </div>
          )}

          <div className="mt-6 pt-6 border-t border-gray-200">
            <p className="text-xs text-gray-500 text-center mb-3">Демо-аккаунты:</p>
            <div className="space-y-2 text-xs text-gray-600">
              <div className="flex justify-between">
                <span>Администратор:</span>
                <code className="bg-gray-100 px-2 py-0.5 rounded">admin / admin123</code>
              </div>
              <div className="flex justify-between">
                <span>Разработчик:</span>
                <code className="bg-gray-100 px-2 py-0.5 rounded">developer / dev123</code>
              </div>
              <div className="flex justify-between">
                <span>QA инженер:</span>
                <code className="bg-gray-100 px-2 py-0.5 rounded">qa / qa123</code>
              </div>
            </div>
          </div>
        </div>
      </div>
    </div>
  );
}
