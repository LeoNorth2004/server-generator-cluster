import { useState, useEffect } from 'react';
import { useTranslation } from 'react-i18next';
import { GlassCard, ChartCard } from '../components/Cards';
import { operationsAPI, clusterAPI } from '../api';
import { useAuth } from '../contexts/AuthContext';
import OperationLogs from '../components/OperationLogs';

export default function Operations() {
  const [loading, setLoading] = useState(false);
  const [metrics, setMetrics] = useState(null);
  const [services, setServices] = useState([]);
  const [systemResources, setSystemResources] = useState(null);
  const { t, i18n } = useTranslation();
  const { user: currentUser } = useAuth();
  const [activeTab, setActiveTab] = useState('monitor'); // 'monitor' or 'logs'
  
  // 自动保活和开关联动
  const [keepAliveEnabled, setKeepAliveEnabled] = useState(false);
  const [autoRefreshEnabled, setAutoRefreshEnabled] = useState(false);
  const [keepAliveInterval, setKeepAliveInterval] = useState(null);
  
  // 开关联动：一个开关动了，另一个也跟着动
  const handleToggleKeepAlive = (enabled) => {
    setKeepAliveEnabled(enabled);
    setAutoRefreshEnabled(enabled); // 联动
    if (enabled) {
      startAutoRefresh();
      recordOperation('keep_alive', 'system', { enabled: true });
    } else {
      stopAutoRefresh();
      recordOperation('keep_alive', 'system', { enabled: false });
    }
  };
  
  const handleToggleAutoRefresh = (enabled) => {
    setAutoRefreshEnabled(enabled);
    setKeepAliveEnabled(enabled); // 联动
    if (enabled) {
      startAutoRefresh();
      recordOperation('refresh', 'system', { auto_refresh: true });
    } else {
      stopAutoRefresh();
      recordOperation('refresh', 'system', { auto_refresh: false });
    }
  };

  const startAutoRefresh = () => {
    if (keepAliveInterval) return;
    const interval = setInterval(() => {
      fetchData();
    }, 30000); // 每30秒自动刷新一次
    setKeepAliveInterval(interval);
  };

  const stopAutoRefresh = () => {
    if (keepAliveInterval) {
      clearInterval(keepAliveInterval);
      setKeepAliveInterval(null);
    }
  };

  // 记录操作的辅助函数
  const recordOperation = async (action, resource, details) => {
    try {
      await operationsAPI.recordOperationLog({
        action,
        resource,
        details: JSON.stringify(details),
        status: 'success'
      });
    } catch (e) {
      console.log('Failed to record operation:', e);
    }
  };

  const fetchData = async () => {
    setLoading(true);
    try {
      // 尝试从集群服务获取真实的系统指标
      let resourcesData = null;
      if (currentUser?.role === 'admin') {
        try {
          const clusterRes = await clusterAPI.getMetrics();
          if (clusterRes.status === 200) {
            resourcesData = clusterRes.data?.data || clusterRes.data || null;
          }
        } catch (e) {
          console.log('Cluster metrics not available');
        }
      }

      // 获取运维数据（传递当前语言）
      const [metricsRes, servicesRes] = await Promise.all([
        operationsAPI.getMetrics(),
        operationsAPI.getServices(),
      ]);

      const metricsData = metricsRes.data?.data || metricsRes.data || {};
      const servicesData = Array.isArray(servicesRes.data?.data) ? servicesRes.data.data :
                          Array.isArray(servicesRes.data) ? servicesRes.data : [];

      // 如果没有从 cluster API 获取到数据，使用 fallback 数据
      if (!resourcesData) {
        resourcesData = {
          total_services: servicesData.length,
          running_services: servicesData.filter(s => s.status === 'Running' || s.status === 'Operational').length,
          services: servicesData.map(s => ({
            name: s.name || s.metadata?.name,
            status: s.status,
            healthy: s.status === 'Running' || s.status === 'Operational',
            cpu_usage: s.cpu_usage || 0,
            memory_usage: s.memory_usage || 0,
          }))
        };
      }

      setMetrics(metricsData);
      setSystemResources(resourcesData);
      if (resourcesData?.services) {
        setServices(resourcesData.services);
      } else {
        setServices(servicesData);
      }
    } catch (error) {
      console.error('Failed to fetch operations data:', error);
      setMetrics(null);
      setServices([]);
      setSystemResources(null);
    } finally {
      setLoading(false);
    }
  };

  const handleRefresh = async () => {
    const startTime = Date.now();
    await fetchData();
    const duration = Date.now() - startTime;
    
    // 记录刷新操作日志
    try {
      await operationsAPI.recordOperationLog({
        action: 'refresh',
        resource: 'operations',
        details: JSON.stringify({ action: 'refresh_monitor', page: 'operations' }),
        status: 'success',
        duration: duration
      });
    } catch (e) {
      console.log('Failed to record refresh log:', e);
    }
  };

  const handleDownload = async () => {
    setLoading(true);
    try {
      const reportContent = generateReport(metrics, services, systemResources);
      
      const blob = new Blob([reportContent], { type: 'text/markdown' });
      const url = URL.createObjectURL(blob);
      const a = document.createElement('a');
      a.href = url;
      a.download = `operations-report-${new Date().toISOString().split('T')[0]}.md`;
      document.body.appendChild(a);
      a.click();
      document.body.removeChild(a);
      URL.revokeObjectURL(url);
      
      alert(t('operations.downloadSuccess'));
    } catch (error) {
      console.error('Failed to download report:', error);
      alert(t('operations.downloadFailed'));
    } finally {
      setLoading(false);
    }
  };

  // 生成报告 - 只使用真实数据，不添加任何虚构内容
  const generateReport = (metrics, services, resources) => {
    let report = `# ${t('operations.title')}\n\n`;
    report += `${t('operations.generatedAt')}: ${new Date().toLocaleString()}\n\n`;
    
    // 系统状态 - 仅显示有数据的部分
    report += `## ${t('operations.systemStatus')}\n`;
    if (metrics && Object.keys(metrics).length > 0) {
      if (metrics.total_requests) report += `- ${t('operations.totalRequests')}: ${Math.round(metrics.total_requests)}\n`;
      if (metrics.avg_response_time) report += `- ${t('operations.avgResponse')}: ${Number(metrics.avg_response_time).toFixed(2)}ms\n`;
    } else {
      report += `- ${t('common.noData')}\n`;
    }
    
    // 系统资源 - 来自 K8s 集群
    report += `\n## ${t('operations.systemResources')}\n`;
    if (resources) {
      if (resources.cpu_usage !== undefined) report += `- ${t('operations.cpuUsage')}: ${resources.cpu_usage}%\n`;
      if (resources.memory_usage !== undefined) report += `- ${t('operations.memory')}: ${resources.memory_usage}%\n`;
      if (resources.disk_usage !== undefined) report += `- ${t('operations.disk')}: ${resources.disk_usage}%\n`;
      if (resources.network_usage !== undefined) report += `- ${t('operations.network')}: ${resources.network_usage}%\n`;
      if (Object.keys(resources).length === 0) report += `- ${t('common.noData')}\n`;
    } else {
      report += `- ${t('common.noData')}\n`;
    }
    
    // 服务状态
    report += `\n## ${t('operations.serviceStatus')}\n`;
    if (services.length > 0) {
      services.forEach(service => {
        const name = service.name || service.metadata?.name;
        const status = service.status || 'Unknown';
        report += `- ${name}: ${status}\n`;
      });
    } else {
      report += `- ${t('common.noData')}\n`;
    }
    
    return report;
  };

  useEffect(() => {
    fetchData();
    
    // 清理定时器
    return () => {
      if (keepAliveInterval) {
        clearInterval(keepAliveInterval);
      }
    };
  }, []);

  // 动态计算系统资源指标（来自 API 或显示暂无数据）
  const getResourceMetrics = () => {
    if (!systemResources) return null;

    if (Array.isArray(systemResources.services)) {
      const services = systemResources.services;
      // 只显示有真实数据的指标（CPU和内存大于0）
      const metrics = [];

      const totalCpu = services.reduce((sum, s) => sum + (s.cpu_usage || 0), 0);
      const totalMem = services.reduce((sum, s) => sum + (s.memory_usage || 0), 0);

      if (totalCpu > 0) {
        const avgCpu = Math.round(totalCpu / services.length);
        metrics.push({
          label: t('operations.cpuUsage'),
          value: avgCpu,
          color: avgCpu > 80 ? 'bg-red-500' : avgCpu > 60 ? 'bg-yellow-500' : 'bg-primary-500'
        });
      }

      if (totalMem > 0) {
        const avgMem = Math.round(totalMem / services.length);
        metrics.push({
          label: t('operations.memory'),
          value: avgMem,
          color: avgMem > 80 ? 'bg-red-500' : avgMem > 60 ? 'bg-yellow-500' : 'bg-purple-500'
        });
      }

      // 磁盘和网络如果没有真实数据就不显示
      if (systemResources.disk_usage && systemResources.disk_usage > 0) {
        metrics.push({
          label: t('operations.disk'),
          value: systemResources.disk_usage,
          color: systemResources.disk_usage > 80 ? 'bg-red-500' : 'bg-blue-500'
        });
      }

      if (systemResources.network_usage && systemResources.network_usage > 0) {
        metrics.push({
          label: t('operations.network'),
          value: systemResources.network_usage,
          color: 'bg-green-500'
        });
      }

      return metrics.length > 0 ? metrics : null;
    }

    const metrics = [];
    if (systemResources.cpu_usage && systemResources.cpu_usage > 0) {
      metrics.push({
        label: t('operations.cpuUsage'),
        value: systemResources.cpu_usage ?? systemResources.cpu ?? null,
        color: systemResources.cpu_usage > 80 ? 'bg-red-5' :
              systemResources.cpu_usage > 60 ? 'bg-yellow-500' : 'bg-primary-500'
      });
    }

    if (systemResources.memory_usage && systemResources.memory_usage > 0) {
      metrics.push({
        label: t('operations.memory'),
        value: systemResources.memory_usage ?? systemResources.memory ?? null,
        color: systemResources.memory_usage > 80 ? 'bg-red-500' :
              systemResources.memory_usage > 60 ? 'bg-yellow-500' : 'bg-purple-500'
      });
    }

    if (systemResources.disk_usage && systemResources.disk_usage > 0) {
      metrics.push({
        label: t('operations.disk'),
        value: systemResources.disk_usage ?? systemResources.disk ?? null,
        color: systemResources.disk_usage > 80 ? 'bg-red-500' :
              systemResources.disk_usage > 60 ? 'bg-yellow-500' : 'bg-blue-500'
      });
    }

    if (systemResources.network_usage && systemResources.network_usage > 0) {
      metrics.push({
        label: t('operations.network'),
        value: systemResources.network_usage ?? null,
        color: 'bg-green-500'
      });
    }

    return metrics.length > 0 ? metrics : null;
  };

  const resourceMetrics = getResourceMetrics();

  // 安全格式化数值显示
  const formatValue = (value, fallback = '-', suffix = '') => {
    if (value === null || value === undefined) return fallback;
    if (typeof value === 'number') {
      return Number(value).toLocaleString() + suffix;
    }
    return String(value) + suffix;
  };

  return (
    <div className="space-y-6">
      <div className="flex items-center justify-between">
        <div>
          <h1 className="text-2xl font-bold text-gray-900 dark:text-white">{t('operations.title')}</h1>
          <p className="text-gray-500 dark:text-gray-400 mt-1">{t('operations.subtitle')}</p>
        </div>
        <div className="flex gap-3">
          {activeTab === 'monitor' && (
            <>
              <button onClick={handleDownload} className="btn-primary flex items-center gap-2" disabled={loading}>
                <svg className="w-5 h-5" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                  <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M4 16v1a3 3 0 003 3h10a3 3 0 003-3v-1m-4-4l-4 4m0 0l-4-4m4 4V4" />
                </svg>
                {t('common.download')}
              </button>
              <button onClick={handleRefresh} className="btn-secondary flex items-center gap-2" disabled={loading}>
                <svg className={`w-5 h-5 ${loading ? 'animate-spin' : ''}`} fill="none" viewBox="0 0 24 24" stroke="currentColor">
                  <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M4 4v5h.582m15.356 2A8.001 8.001 0 004.582 9m0 0H9m11 11v-5h-.581m0 0a8.003 8.003 0 01-15.357-2m15.357 2H15" />
                </svg>
                {loading ? t('common.loading') : t('common.refresh')}
              </button>
            </>
          )}
        </div>
      </div>

      {/* Tab 切换 */}
      <div className="flex space-x-1 bg-gray-100 dark:bg-gray-800 p-1 rounded-lg w-fit">
        <button
          onClick={() => setActiveTab('monitor')}
          className={`px-4 py-2 text-sm font-medium rounded-md transition-colors ${
            activeTab === 'monitor'
              ? 'bg-white dark:bg-gray-700 text-primary-600 shadow-sm'
              : 'text-gray-600 dark:text-gray-400 hover:text-gray-900 dark:hover:text-white'
          }`}
        >
          {t('operations.systemMetrics')}
        </button>
        <button
          onClick={() => setActiveTab('logs')}
          className={`px-4 py-2 text-sm font-medium rounded-md transition-colors ${
            activeTab === 'logs'
              ? 'bg-white dark:bg-gray-700 text-primary-600 shadow-sm'
              : 'text-gray-600 dark:text-gray-400 hover:text-gray-900 dark:hover:text-white'
          }`}
        >
          {t('operations.operationLogs')}
        </button>
      </div>

      {/* Tab 内容 */}
      {activeTab === 'monitor' ? (
        <>
        {/* 统计卡片 */}
        <div className="grid grid-cols-1 md:grid-cols-2 gap-6">
          <div className="bg-gradient-to-br from-blue-50 to-indigo-100 dark:from-blue-900/30 dark:to-indigo-900/30 rounded-2xl p-6 border border-blue-200 dark:border-blue-800">
            <div style={{display:'flex', justifyContent:'space-between', alignItems:'center', marginBottom:'12px'}}>
              <span className="text-sm font-medium text-blue-700 dark:text-blue-300">{t('operations.totalRequests')}</span>
              <span style={{fontSize:'20px'}}>📊</span>
            </div>
            <p className="text-4xl font-bold text-blue-900 dark:text-blue-100">{formatValue(metrics?.total_requests, '0', '')}</p>
            <p className="text-xs text-blue-600 dark:text-blue-400 mt-2">累计 API 请求总数</p>
          </div>

          <div className="bg-gradient-to-br from-emerald-50 to-teal-100 dark:from-emerald-900/30 dark:to-teal-900/30 rounded-2xl p-6 border border-emerald-200 dark:border-emerald-800">
            <div style={{display:'flex', justifyContent:'space-between', alignItems:'center', marginBottom:'12px'}}>
              <span className="text-sm font-medium text-emerald-700 dark:text-emerald-300">{t('operations.avgResponse')}</span>
              <span style={{fontSize:'20px'}}>⚡</span>
            </div>
            <p className="text-4xl font-bold text-emerald-900 dark:text-emerald-100">{formatValue(metrics?.avg_response_time, '0', 'ms')}</p>
            <p className="text-xs text-emerald-600 dark:text-emerald-400 mt-2">平均 API 响应时间</p>
          </div>
        </div>

      <div className="grid grid-cols-1 lg:grid-cols-2 gap-6">
        {/* 系统资源指标 - 仅在有真实数据时显示整个组件 */}
        {resourceMetrics && resourceMetrics.length > 0 && (
          <ChartCard title={t('operations.systemMetrics')} className="lg:col-span-2">
            <div className="grid grid-cols-2 gap-6">
              {resourceMetrics.map((metric, index) => (
                <div key={index} className="space-y-2">
                  <div className="flex justify-between">
                    <span className="text-sm font-medium text-gray-700 dark:text-gray-300">{metric.label}</span>
                    <span className="text-sm text-gray-500">{metric.value}%</span>
                  </div>
                  <div className="h-2 bg-gray-200 dark:bg-gray-700 rounded-full overflow-hidden">
                    <div
                      className={`h-full ${metric.color} rounded-full transition-all duration-500`}
                      style={{ width: `${Math.min(metric.value, 100)}%` }}
                    />
                  </div>
                </div>
              ))}
            </div>
          </ChartCard>
        )}

        {/* 服务健康状态 - 来自真实 API */}
        <GlassCard>
          <h3 className="text-lg font-semibold text-gray-900 dark:text-white mb-4">{t('operations.serviceHealth')}</h3>
          <div className="space-y-4">
            {Array.isArray(services) && services.length > 0 ? (
              services.map((service, index) => {
                const serviceName = service.name || service.metadata?.name || `service-${index}`;
                const isOperational = service.status === 'Operational' || service.status === 'Running' || service.status === 'Healthy';
                
                return (
                  <div key={index} className="flex items-center justify-between">
                    <div className="flex items-center gap-2">
                      <span className={`w-2 h-2 rounded-full ${isOperational ? 'bg-green-500 animate-pulse' : 'bg-yellow-500'}`} />
                      <span className="text-sm font-medium text-gray-700 dark:text-gray-300">{serviceName}</span>
                    </div>
                    <span className="text-sm text-gray-500">
                      {service.latency || service.uptime || service.response_time || '-'}
                    </span>
                  </div>
                );
              })
            ) : (
              <p className="text-gray-500">{t('common.noData')}</p>
            )}
          </div>
        </GlassCard>
      </div>

      <div className="grid grid-cols-1 lg:grid-cols-2 gap-6">
        {/* 服务详细状态 */}
        <ChartCard title={t('operations.serviceStatus')}>
          <div className="space-y-4">
            {Array.isArray(services) && services.length > 0 ? (
              services.map((service, index) => {
                const serviceName = service.name || service.metadata?.name || `service-${index}`;
                const status = service.status || 'Unknown';
                const uptime = service.uptime || service.ready || '-';
                const isOperational = status === 'Operational' || status === 'Running' || status === 'Healthy';
                
                return (
                  <div key={index} className="flex items-center justify-between p-3 rounded-xl bg-gray-50 dark:bg-gray-800/50">
                    <div className="flex items-center gap-3">
                      <span className={`w-3 h-3 rounded-full ${isOperational ? 'bg-green-500' : 'bg-yellow-500'}`} />
                      <span className="font-medium text-gray-900 dark:text-white">{serviceName}</span>
                    </div>
                    <div className="flex items-center gap-4">
                      <span className="text-sm text-gray-500">{uptime}</span>
                      <span className={`px-2 py-1 text-xs font-medium rounded-full ${
                        isOperational
                          ? 'bg-green-100 text-green-700 dark:bg-green-900/30 dark:text-green-400'
                          : 'bg-yellow-100 text-yellow-700 dark:bg-yellow-900/30 dark:text-yellow-400'
                      }`}>
                        {status}
                      </span>
                    </div>
                  </div>
                );
              })
            ) : (
              <p className="text-gray-500">{t('common.noData')}</p>
            )}
          </div>
        </ChartCard>
      </div>
        </>
      ) : (
        <OperationLogs />
      )}
    </div>
  );
}
