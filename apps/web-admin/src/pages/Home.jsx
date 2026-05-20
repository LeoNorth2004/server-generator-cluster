import { useState, useEffect } from 'react';
import { useNavigate } from 'react-router-dom';
import { useTranslation } from 'react-i18next';
import { StatCard, ChartCard, GlassCard } from '../components/Cards';
import { projectAPI, userAPI, clusterAPI, generatorAPI } from '../api';

export default function Home() {
  const [stats, setStats] = useState({ user_count: 0, project_count: 0, generation_count: 0 });
  const [health, setHealth] = useState(null);
  const [loading, setLoading] = useState(false);
  const [recentProjects, setRecentProjects] = useState([]);
  const [clusterStatus, setClusterStatus] = useState(null);
  const [services, setServices] = useState([]);
  const { t } = useTranslation();
  const navigate = useNavigate();

  useEffect(() => {
    const fetchData = async () => {
      setLoading(true);
      try {
        // 并行获取所有真实数据
        const [usersRes, projectsRes, healthRes, clusterRes, servicesRes] = await Promise.allSettled([
          userAPI.list(),
          projectAPI.list(),
          fetch('/api/v1/operations/health'),
          clusterAPI.getK8sStatus(),
          clusterAPI.getK8sServices('generator-platform').catch(() => null),
        ]);

        // 用户统计
        if (usersRes.status === 'fulfilled' && usersRes.value.status === 200) {
          const usersData = usersRes.value.data?.data || usersRes.value.data || [];
          setStats(prev => ({ ...prev, user_count: Array.isArray(usersData) ? usersData.length : 0 }));
        }

        // 项目统计
        if (projectsRes.status === 'fulfilled' && projectsRes.value.status === 200) {
          const projectsData = projectsRes.value.data?.data || projectsRes.value.data || [];
          setStats(prev => ({ ...prev, project_count: Array.isArray(projectsData) ? projectsData.length : 0 }));
          
          // 最近项目（取前5个）
          if (Array.isArray(projectsData)) {
            setRecentProjects(projectsData.slice(0, 5));
          }
        }

        // 健康检查
        if (healthRes.status === 'fulfilled' && healthRes.value.ok) {
          const healthData = await healthRes.value.json();
          setHealth(healthData);
        }

        // 集群状态
        if (clusterRes.status === 'fulfilled' && clusterRes.value.status === 200) {
          const raw = clusterRes.value.data;
          const clusterData = raw?.data ?? raw;
          setClusterStatus(clusterData);
        } else {
          console.log('[DEBUG] Cluster API failed:', clusterRes.status, clusterRes.reason?.message);
        }

        // 服务列表（用于系统健康显示）
        if (servicesRes.status === 'fulfilled' && servicesRes.value?.status === 200) {
          const servicesData = servicesRes.value.data?.data || [];
          if (Array.isArray(servicesData)) {
            setServices(servicesData.slice(0, 6)); // 只显示前6个
          }
        }
      } catch (error) {
        console.log('Some data not available:', error.message);
      } finally {
        setLoading(false);
      }
    };
    fetchData();
  }, []);

  const handleNewProject = () => navigate('/generator');
  const handleGenerateCode = () => navigate('/generator');
  const handleManageClusters = () => navigate('/clusters');

  const statCards = [
    {
      title: t('dashboard.totalUsers'),
      value: stats.user_count,
      icon: (
        <svg className="w-6 h-6" fill="none" viewBox="0 0 24 24" stroke="currentColor">
          <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M12 4.354a4 4 0 110 5.292M15 21H3v-1a6 6 0 0112 0v1zm0 0h6v-1a6 6 0 00-9-5.197M13 7a4 4 0 11-8 0 4 4 0 018 0z" />
        </svg>
      ),
    },
    {
      title: t('dashboard.activeProjects'),
      value: stats.project_count,
      icon: (
        <svg className="w-6 h-6" fill="none" viewBox="0 0 24 24" stroke="currentColor">
          <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M3 7v10a2 2 0 002 2h14a2 2 0 002-2V9a2 2 0 00-2-2h-6l-2-2H5a2 2 0 00-2 2z" />
        </svg>
      ),
    },
    {
      title: t('dashboard.codeGenerations'),
      value: stats.generation_count,
      icon: (
        <svg className="w-6 h-6" fill="none" viewBox="0 0 24 24" stroke="currentColor">
          <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M10 20l4-16m4 4l4 4-4 4M6 16l-4-4 4-4" />
        </svg>
      ),
    },
    {
      title: t('dashboard.clusterStatus'),
      value: clusterStatus?.connected ? t('dashboard.healthy') : t('dashboard.unknown'),
      icon: (
        <svg className="w-6 h-6" fill="none" viewBox="0 0 24 24" stroke="currentColor">
          <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M5 12h14M5 12a2 2 0 01-2-2V6a2 2 0 012-2h14a2 2 0 012 2v4a2 2 0 01-2 2M5 12a2 2 0 00-2 2v4a2 2 0 002 2h14a2 2 0 002-2v-4a2 2 0 00-2-2m-2-4h.01M17 16h.01" />
        </svg>
      ),
    },
  ];

  return (
    <div className="space-y-6">
      <div className="flex items-center justify-between">
        <div>
          <h1 className="text-2xl font-bold text-gray-900 dark:text-white">{t('dashboard.title')}</h1>
          <p className="text-gray-500 dark:text-gray-400 mt-1">{t('dashboard.welcome')} - Generator Platform</p>
        </div>
        <div className="flex gap-3">
          <button onClick={() => alert(t('dashboard.downloadReportDeveloping'))} className="btn-secondary">{t('dashboard.downloadReport')}</button>
          <button onClick={handleNewProject} className="btn-primary">{t('dashboard.newProject')}</button>
        </div>
      </div>

      {/* 统计卡片 */}
      <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-4 gap-6">
        {statCards.map((stat, index) => (
          <StatCard key={index} {...stat} />
        ))}
      </div>

      <div className="grid grid-cols-1 lg:grid-cols-3 gap-6">
        {/* 快速操作 */}
        <GlassCard>
          <h3 className="text-lg font-semibold text-gray-900 dark:text-white mb-4">{t('dashboard.quickActions')}</h3>
          <div className="space-y-3">
            <button onClick={handleNewProject} className="w-full btn-primary flex items-center justify-center gap-2">
              <svg className="w-5 h-5" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M12 4v16m8-8H4" />
              </svg>
              {t('dashboard.newProject')}
            </button>
            <button onClick={handleGenerateCode} className="w-full btn-secondary flex items-center justify-center gap-2">
              <svg className="w-5 h-5" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M10 20l4-16m4 4l4 4-4 4M6 16l-4-4 4-4" />
              </svg>
              {t('dashboard.generateCode')}
            </button>
            <button onClick={handleManageClusters} className="w-full btn-secondary flex items-center justify-center gap-2">
              <svg className="w-5 h-5" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M5 12h14M5 12a2 2 0 01-2-2V6a2 2 0 012-2h14a2 2 0 012 2v4a2 2 0 01-2 2" />
              </svg>
              {t('dashboard.manageClusters')}
            </button>
          </div>
        </GlassCard>

        {/* 最近项目 - 来自真实 API */}
        <ChartCard title={t('dashboard.recentProjects')} className="lg:col-span-2">
          <div className="space-y-4 max-h-[320px] overflow-y-auto">
            {recentProjects.length > 0 ? recentProjects.map((project, index) => (
              <div key={index} className="flex items-center justify-between p-3 rounded-xl bg-gray-50 dark:bg-gray-800/50 hover:bg-gray-100 dark:hover:bg-gray-800 transition-colors">
                <div className="flex items-center gap-3">
                  <div className="w-10 h-10 rounded-lg bg-gradient-to-br from-primary-500 to-primary-700 flex items-center justify-center">
                    <svg className="w-5 h-5 text-white" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                      <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M3 7v10a2 2 0 002 2h14a2 2 0 002-2V9a2 2 0 00-2-2h-6l-2-2H5a2 2 0 00-2 2z" />
                    </svg>
                  </div>
                  <div>
                    <p className="font-medium text-gray-900 dark:text-white">{project.name}</p>
                    <p className="text-sm text-gray-500 dark:text-gray-400">
                      {project.created_at 
                        ? new Date(project.created_at).toLocaleDateString() 
                        : project.updated_at 
                          ? new Date(project.updated_at).toLocaleDateString()
                          : '-'
                      }
                    </p>
                  </div>
                </div>
                <span className={`px-3 py-1 text-xs font-medium rounded-full ${
                  project.generated_code 
                    ? 'bg-green-100 text-green-700 dark:bg-green-900/30 dark:text-green-400'
                    : 'bg-yellow-100 text-yellow-700 dark:bg-yellow-900/30 dark:text-yellow-400'
                }`}>
                  {project.generated_code ? t('dashboard.generated') : t('dashboard.notGenerated')}
                </span>
              </div>
            )) : (
              <div className="flex flex-col items-center justify-center py-12 text-gray-400">
                <svg className="w-12 h-12 mb-3 opacity-50" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                  <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M3 7v10a2 2 0 002 2h14a2 2 0 002-2V9a2 2 0 00-2-2h-6l-2-2H5a2 2 0 00-2 2z" />
                </svg>
                <p className="font-medium">{t('common.noData')}</p>
                <p className="text-sm mt-1">{t('dashboard.noProjectsYet')}</p>
              </div>
            )}
          </div>
        </ChartCard>
      </div>
    </div>
  );
}
