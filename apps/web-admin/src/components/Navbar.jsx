import { useTheme } from '../contexts/ThemeContext';
import { useAuth } from '../contexts/AuthContext';
import { useI18n } from '../contexts/I18nContext';

export default function Navbar() {
  const { isDark, toggleTheme } = useTheme();
  const { user, logout } = useAuth();
  const { language, toggleLanguage } = useI18n();

  return (
    <nav
      className="fixed top-0 left-0 right-0 h-16 z-50 flex items-center px-6"
      style={{ backgroundColor: 'var(--sidebar-bg)', borderBottom: '1px solid var(--border-color)' }}
    >
      <div className="flex items-center gap-3">
        <div className="w-8 h-8 rounded-lg bg-gradient-to-br from-primary-500 to-primary-700 flex items-center justify-center">
          <svg className="w-5 h-5 text-white" fill="none" viewBox="0 0 24 24" stroke="currentColor">
            <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M13 10V3L4 14h7v7l9-11h-7z" />
          </svg>
        </div>
        <span className="text-lg font-semibold" style={{ color: 'var(--text-primary)' }}>Generator Platform</span>
      </div>

      <div className="flex items-center gap-4 ml-auto">
        <button
          onClick={toggleLanguage}
          className="px-3 py-1.5 rounded-lg text-sm font-medium transition-colors"
          style={{
            backgroundColor: 'rgba(147, 51, 234, 0.1)',
            color: 'rgb(147 51 234)'
          }}
        >
          {language === 'zh' ? 'EN' : '中'}
        </button>

        <button
          onClick={toggleTheme}
          className="p-2 rounded-xl transition-colors"
          style={{ backgroundColor: 'var(--hover-bg)' }}
        >
          {isDark ? (
            <svg className="w-5 h-5 text-yellow-500" fill="none" viewBox="0 0 24 24" stroke="currentColor">
              <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M12 3v1m0 16v1m9-9h-1M4 12H3m15.364 6.364l-.707-.707M6.343 6.343l-.707-.707m12.728 0l-.707.707M6.343 17.657l-.707.707M16 12a4 4 0 11-8 0 4 4 0 018 0z" />
            </svg>
          ) : (
            <svg className="w-5 h-5" style={{ color: 'var(--text-secondary)' }} fill="none" viewBox="0 0 24 24" stroke="currentColor">
              <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M20.354 15.354A9 9 0 018.646 3.646 9.003 9.003 0 0012 21a9.003 9.003 0 008.354-5.646z" />
            </svg>
          )}
        </button>

        <div className="flex items-center gap-3 pl-4" style={{ borderLeft: '1px solid var(--border-color)' }}>
          <div className="w-8 h-8 rounded-full bg-gradient-to-br from-primary-400 to-primary-600 flex items-center justify-center">
            <span className="text-sm font-medium text-white">{user?.username?.charAt(0).toUpperCase() || 'U'}</span>
          </div>
          <div className="hidden md:block">
            <p className="text-sm font-medium" style={{ color: 'var(--text-primary)' }}>{user?.username || 'User'}</p>
            <p className="text-xs" style={{ color: 'var(--text-muted)' }}>{user?.role || 'user'}</p>
          </div>
          <button
            onClick={logout}
            className="ml-2 p-2 rounded-xl transition-colors"
            style={{ backgroundColor: 'var(--hover-bg)' }}
          >
            <svg className="w-5 h-5" style={{ color: 'var(--text-muted)' }} fill="none" viewBox="0 0 24 24" stroke="currentColor">
              <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M17 16l4-4m0 0l-4-4m4 4H7m6 4v1a3 3 0 01-3 3H6a3 3 0 01-3-3V7a3 3 0 013-3h4a3 3 0 013 3v1" />
            </svg>
          </button>
        </div>
      </div>
    </nav>
  );
}