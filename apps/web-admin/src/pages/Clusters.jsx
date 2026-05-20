import { useState, useEffect, useCallback } from 'react';
import { useTranslation } from 'react-i18next';
import { GlassCard, StatCard, ChartCard } from '../components/Cards';
import { clusterAPI } from '../api';

export default function Clusters() {
  const [activeTab, setActiveTab] = useState('overview');
  const [clusterInfo, setClusterInfo] = useState({
    k8s: { connected: false, inCluster: false, namespace: '', version: '', error: '' },
    stats: { totalPods: 0, runningPods: 0, totalServices: 0, totalNodes: 0 }
  });
  const [namespaces, setNamespaces] = useState([]);
  const [nodes, setNodes] = useState([]);
  const [pods, setPods] = useState([]);
  const [services, setServices] = useState([]);
  const [deployments, setDeployments] = useState([]);
  const [selectedNamespace, setSelectedNamespace] = useState('generator-platform');
  const [loading, setLoading] = useState(false);
  const [showLogModal, setShowLogModal] = useState(false);
  const [selectedPod, setSelectedPod] = useState(null);
  const [podLogs, setPodLogs] = useState('');
  const [showScaleModal, setShowScaleModal] = useState(false);
  const [selectedDeployment, setSelectedDeployment] = useState(null);
  const [scaleReplicas, setScaleReplicas] = useState(1);
  const [autoKeepalive, setAutoKeepalive] = useState(false);
  const [keepaliveInterval, setKeepaliveInterval] = useState(30);
  const [healthLog, setHealthLog] = useState([]);
  const [lastHealthCheck, setLastHealthCheck] = useState(null);
  const [healthStats, setHealthStats] = useState({ healthy: 0, unhealthy: 0, total: 0 });
  const [nodeScalingInfo, setNodeScalingInfo] = useState(null);
  const [detailedNodes, setDetailedNodes] = useState([]);
  const [showAddNodeModal, setShowAddNodeModal] = useState(false);
  const [scalingLoading, setScalingLoading] = useState(false);
  
  // ===== 自动保活状态 =====
  const [autoHealingStatus, setAutoHealingStatus] = useState(null);
  const [autoHealingConfig, setAutoHealingConfig] = useState({ enabled: true, interval_seconds: 30, max_auto_nodes: 5 });
  const [healingHistory, setHealingHistory] = useState({ memory_events: [], db_logs: [] });
  const { t } = useTranslation();

  const fetchClusterStatus = useCallback(async () => {
    try {
      setLoading(true);
      const response = await clusterAPI.getK8sStatus();
      if (response.status === 200) {
        const data = response.data?.data || response.data || {};
        setClusterInfo(prev => ({
          ...prev,
          k8s: {
            connected: data.connected || false,
            inCluster: data.in_cluster || false,
            namespace: data.namespace || '',
            version: data.version || '',
            error: data.error || '',
            mode: data.mode || '',
            mode_display: data.mode_display || ''
          }
        }));

        if (Array.isArray(data.namespaces)) {
          setNamespaces(data.namespaces);
        }
        if (Array.isArray(data.nodes)) {
          setNodes(data.nodes);
        }
      }
    } catch (error) {
      console.error('Failed to fetch cluster status:', error);
      setClusterInfo(prev => ({
        ...prev,
        k8s: { connected: false, inCluster: false, namespace: '', version: '', error: error.message, mode: '', mode_display: '' }
      }));
    } finally {
      setLoading(false);
    }
  }, []);

  const fetchPods = useCallback(async (namespace) => {
    try {
      const response = await clusterAPI.getK8sPods(namespace || selectedNamespace);
      if (response.status === 200) {
        const data = response.data?.data || response.data || [];
        setPods(Array.isArray(data) ? data : []);
      }
    } catch (error) {
      console.error('Failed to fetch pods:', error);
      setPods([]);
    }
  }, [selectedNamespace]);

  const fetchServices = useCallback(async () => {
    try {
      const response = await clusterAPI.getK8sServices(selectedNamespace);
      if (response.status === 200) {
        const data = response.data?.data || response.data || [];
        setServices(Array.isArray(data) ? data : []);
      }
    } catch (error) {
      console.error('Failed to fetch services:', error);
      setServices([]);
    }
  }, [selectedNamespace]);

  const fetchNodes = useCallback(async () => {
    try {
      const response = await clusterAPI.getK8sNodes();
      if (response.status === 200) {
        const data = response.data?.data || response.data || [];
        setNodes(Array.isArray(data) ? data : []);
      }
    } catch (error) {
      console.error('Failed to fetch nodes:', error);
      setNodes([]);
    }
  }, []);

  const fetchDeployments = useCallback(async () => {
    try {
      const response = await clusterAPI.getK8sDeployments(selectedNamespace);
      if (response.status === 200) {
        const data = response.data?.data || response.data || [];
        setDeployments(Array.isArray(data) ? data : []);
      }
    } catch (error) {
      console.error('Failed to fetch deployments:', error);
      setDeployments([]);
    }
  }, [selectedNamespace]);

  useEffect(() => {
    fetchClusterStatus();
  }, [fetchClusterStatus]);

  useEffect(() => {
    if (clusterInfo.k8s.connected) {
      fetchNodes();
      fetchPods();
      fetchServices();
      fetchDeployments();
    }
  }, [clusterInfo.k8s.connected, selectedNamespace, activeTab, fetchNodes, fetchPods, fetchServices, fetchDeployments]);

  const handleViewPodLogs = async (pod) => {
    setSelectedPod(pod);
    setShowLogModal(true);
    setPodLogs(t('clusters.loadingLogs'));
    
    try {
      const response = await clusterAPI.getK8sPodLogs(pod.namespace || selectedNamespace, pod.name);
      if (response.status === 200) {
        const data = response.data?.data || response.data || {};
        setPodLogs(data.logs || t('clusters.noLogs'));
      }
    } catch (error) {
      setPodLogs(`${t('clusters.failedToLoadLogs')}: ${error.message}`);
    }
  };

  const handleDeletePod = async (pod) => {
    if (!window.confirm(`${t('clusters.confirmDeletePod')} ${pod.name}?`)) return;
    
    try {
      await clusterAPI.deleteK8sPod(pod.namespace || selectedNamespace, pod.name);
      alert(t('clusters.podDeleted'));
      fetchPods();
    } catch (error) {
      alert(`${t('clusters.failedToDeletePod')}: ${error.message}`);
    }
  };

  const handleRestartDeployment = async (deployment) => {
    try {
      await clusterAPI.restartK8sDeployment(deployment.namespace || selectedNamespace, deployment.name);
      alert(t('clusters.deploymentRestarted'));
      fetchPods();
      fetchDeployments();
    } catch (error) {
      alert(`${t('clusters.failedToRestart')}: ${error.message}`);
    }
  };

  const handleScaleDeployment = async () => {
    if (!selectedDeployment) return;

    try {
      await clusterAPI.scaleK8sDeployment(
        selectedDeployment.namespace || selectedNamespace,
        selectedDeployment.name,
        parseInt(scaleReplicas)
      );
      alert(t('clusters.deploymentScaled'));
      setShowScaleModal(false);
      fetchDeployments();
      fetchPods();
    } catch (error) {
      alert(`${t('clusters.failedToScale')}: ${error.message}`);
    }
  };

  const runHealthCheck = async () => {
    const timestamp = new Date().toLocaleTimeString();
    let allHealthy = true;
    let issues = [];

    try {
      const podsRes = await clusterAPI.getK8sPods(selectedNamespace);
      if (podsRes.status === 200) {
        const podsData = podsRes.data?.data || podsRes.data || [];
        const podList = Array.isArray(podsData) ? podsData : [];
        let healthyCount = 0;
        let unhealthyCount = 0;

        podList.forEach(pod => {
          const podStatus = pod.status || (pod.status && pod.status.phase) || 'Unknown';
          if (podStatus === 'Running' || podStatus === 'Succeeded') {
            healthyCount++;
          } else {
            unhealthyCount++;
            allHealthy = false;
            issues.push(`Pod ${pod.name || 'unknown'}: ${podStatus}`);
          }
        });

        setHealthStats({ healthy: healthyCount, unhealthy: unhealthyCount, total: podList.length });
      }
    } catch (e) {
      issues.push(`Pod check failed: ${e.message}`);
      allHealthy = false;
    }

    const entry = {
      time: timestamp,
      status: allHealthy ? 'healthy' : 'warning',
      message: allHealthy
        ? `${healthStats.total || 0} pods running normally`
        : `Found ${issues.length} issue(s): ${issues.join('; ')}`,
      details: issues
    };

    setHealthLog(prev => [entry, ...prev].slice(0, 50));
    setLastHealthCheck(timestamp);

    if (!allHealthy && autoKeepalive) {
      setHealthLog(prev => [...prev, {
        time: new Date().toLocaleTimeString(),
        status: 'info',
        message: 'Auto-keepalive: Attempting to restart unhealthy deployments...'
      }]);
      
      if (Array.isArray(deployments)) {
        deployments.forEach(dep => {
          if (dep.available_replicas !== dep.replicas) {
            handleRestartDeployment(dep).catch(() => {});
          }
        });
      }
    }

    return entry;
  };

  useEffect(() => {
    let interval = null;
    if (autoKeepalive && clusterInfo.k8s.connected) {
      interval = setInterval(() => {
        runHealthCheck();
        fetchPods();
        fetchDeployments();
      }, keepaliveInterval * 1000);
    }
    return () => { if (interval) clearInterval(interval); };
  }, [autoKeepalive, keepaliveInterval, clusterInfo.k8s.connected]);

  // ===== 节点扩缩容功能 =====

  const fetchNodeScalingInfo = useCallback(async () => {
    try {
      const response = await clusterAPI.getNodeScalingInfo();
      if (response.status === 200) {
        const data = response.data?.data || response.data || {};
        setNodeScalingInfo(data);
      }
    } catch (error) {
      console.error('Failed to fetch node scaling info:', error);
    }
  }, []);

  const fetchNodesDetailed = useCallback(async () => {
    try {
      const response = await clusterAPI.getNodesDetailed();
      if (response.status === 200) {
        const data = response.data?.data || response.data || {};
        setDetailedNodes(data.nodes || data || []);
      }
    } catch (error) {
      console.error('Failed to fetch detailed nodes:', error);
    }
  }, []);

  useEffect(() => {
    if (activeTab === 'nodes' && clusterInfo.k8s.connected) {
      fetchNodeScalingInfo();
      fetchNodesDetailed();
    }
  }, [activeTab, clusterInfo.k8s.connected]);

  const handleAddNode = async (role = 'agent') => {
    if (!window.confirm(t('clusters.confirmAddNode', { role }))) return;

    setScalingLoading(true);
    try {
      const response = await clusterAPI.scaleNode({
        action: 'add',
        role: role,
      });

      if (response.status === 200) {
        const result = response.data?.data || response.data || {};
        alert(`✅ ${result.message || t('clusters.nodeAddedSuccess')}`);
        setShowAddNodeModal(false);
        setTimeout(() => {
          fetchNodes();
          fetchNodeScalingInfo();
          fetchNodesDetailed();
        }, 3000);
      } else {
        alert(`❌ ${t('clusters.addNodeFailed')}: ${response.data?.message || t('clusters.unknownError')}`);
      }
    } catch (error) {
      alert(`❌ ${t('clusters.addNodeFailedMsg')}: ${error.message}`);
    } finally {
      setScalingLoading(false);
    }
  };

  const handleRestartNode = async (nodeName) => {
    if (!window.confirm(t('clusters.confirmRestartNode', { nodeName }))) return;

    setScalingLoading(true);
    try {
      const response = await clusterAPI.scaleNode({
        action: 'restart',
        node_name: nodeName,
      });

      if (response.status === 200) {
        const result = response.data?.data || response.data || {};
        alert(`✅ ${result.message || t('clusters.nodeRestartSuccess')}`);
        setTimeout(() => {
          fetchNodes();
          fetchNodesDetailed();
        }, 2000);
      } else {
        alert(`❌ ${t('clusters.restartFailed')}: ${response.data?.message || t('clusters.unknownError')}`);
      }
    } catch (error) {
      alert(`❌ ${t('clusters.restartNodeFailedMsg')}: ${error.message}`);
    } finally {
      setScalingLoading(false);
    }
  };

  // ===== 自动保活功能 =====

  const fetchAutoHealingStatus = useCallback(async () => {
    try {
      const response = await clusterAPI.getAutoHealingStatus();
      if (response.status === 200) {
        const data = response.data?.data || response.data || {};
        setAutoHealingStatus(data);
        if (data.config) {
          setAutoHealingConfig(data.config);
        }
      }
    } catch (error) {
      console.error('Failed to fetch auto-healing status:', error);
    }
  }, []);

  const fetchHealingHistory = useCallback(async () => {
    try {
      const response = await clusterAPI.getHealingHistory();
      if (response.status === 200) {
        const data = response.data?.data || response.data || {};
        setHealingHistory({
          memory_events: data.memory_events || [],
          db_logs: data.db_logs || [],
        });
      }
    } catch (error) {
      console.error('Failed to fetch healing history:', error);
    }
  }, []);

  useEffect(() => {
    if (activeTab === 'nodes' && clusterInfo.k8s.connected) {
      fetchAutoHealingStatus();
      fetchHealingHistory();
      
      // 每15秒刷新一次状态
      const interval = setInterval(() => {
        fetchAutoHealingStatus();
        fetchHealingHistory();
      }, 15000);
      return () => clearInterval(interval);
    }
  }, [activeTab, clusterInfo.k8s.connected]);

  const handleToggleAutoHealing = async () => {
    const newConfig = { ...autoHealingConfig, enabled: !autoHealingConfig.enabled };

    try {
      await clusterAPI.updateAutoHealingConfig(newConfig);
      setAutoHealingConfig(newConfig);
      alert(t('clusters.keepaliveStatusChanged', { status: newConfig.enabled ? '启用 ✅' : '禁用 ❌' }));
      fetchAutoHealingStatus();
    } catch (error) {
      alert(`${t('clusters.operationFailed')}: ${error.message}`);
    }
  };

  const handleUpdateHealingConfig = async () => {
    try {
      await clusterAPI.updateAutoHealingConfig(autoHealingConfig);
      alert(t('clusters.configUpdated'));
      fetchAutoHealingStatus();
    } catch (error) {
      alert(`${t('clusters.updateFailedMsg')}: ${error.message}`);
    }
  };

  const handleTriggerManualCheck = async () => {
    try {
      await clusterAPI.triggerManualHealthCheck();
      alert(t('clusters.healthCheckTriggered'));
      setTimeout(() => {
        fetchAutoHealingStatus();
        fetchHealingHistory();
      }, 5000);
    } catch (error) {
      alert(`${t('clusters.triggerFailed')}: ${error.message}`);
    }
  };

  const tabs = [
    { id: 'overview', label: t('clusters.platformOverview'), icon: '📊' },
    { id: 'services', label: t('clusters.microservices'), icon: '🔧' },
    { id: 'pods', label: t('clusters.pods'), icon: '📦' },
    { id: 'deployments', label: t('clusters.deployments'), icon: '🚀' },
    { id: 'nodes', label: t('clusters.nodes'), icon: '🖥️' },
    { id: 'health', label: t('clusters.healthCheck'), icon: '✨' },
  ];

  return (
    <div className="space-y-6">
      <div className="flex items-center justify-between">
        <div>
          <h1 className="text-2xl font-bold text-gray-900 dark:text-white">{t('clusters.title')}</h1>
          <p className="text-gray-500 dark:text-gray-400 mt-1">{t('clusters.subtitlePlatform')}</p>
        </div>
        <button 
          onClick={fetchClusterStatus} 
          className="btn-secondary"
          disabled={loading}
        >
          {loading ? '⟳' : '🔄'} {t('common.refresh')}
        </button>
      </div>

      <div className="grid grid-cols-1 md:grid-cols-4 gap-6">
        <StatCard
          title={t('clusters.k8sStatus')}
          value={clusterInfo.k8s.connected ? (clusterInfo.k8s.mode === 'docker' ? '🐳 Docker' : '☸️ K8s') : t('clusters.disconnected')}
          icon={<span className={`w-3 h-3 rounded-full ${clusterInfo.k8s.connected ? 'bg-green-500 animate-pulse' : 'bg-red-500'}`} />}
          subtitle={clusterInfo.k8s.mode_display || null}
        />
        <StatCard
          title={t('clusters.nodes')}
          value={nodes.length}
          icon={<span>🖥️</span>}
          trend={null}
        />
        <StatCard
          title={t('clusters.pods')}
          value={`${pods.filter(p => p.status === 'Running').length}/${pods.length}`}
          icon={<span>📦</span>}
          trend={null}
        />
        <StatCard
          title={t('clusters.services')}
          value={services.length}
          icon={<span>🔗</span>}
          trend={null}
        />
      </div>

      {!clusterInfo.k8s.connected && (
        <GlassCard>
          <div className="text-center py-12">
            <div className="text-6xl mb-4">⚠️</div>
            <h3 className="text-xl font-semibold text-gray-900 dark:text-white mb-2">
              {t('clusters.k8sNotConnected')}
            </h3>
            <p className="text-gray-500 mb-4 max-w-md mx-auto">
              {clusterInfo.k8s.error || t('clusters.k8sNotConnectedDesc')}
            </p>
            <div className="bg-gray-50 dark:bg-gray-800 rounded-lg p-4 max-w-lg mx-auto text-left text-sm">
              <p className="font-medium mb-2">{t('clusters.howToConnect')}:</p>
              <ol className="list-decimal list-inside space-y-1 text-gray-600 dark:text-gray-400">
                <li>{t('clusters.step1')}</li>
                <li>{t('clusters.step2')}</li>
                <li>{t('clusters.step3')}</li>
              </ol>
            </div>
          </div>
        </GlassCard>
      )}

      {clusterInfo.k8s.connected && clusterInfo.k8s.mode === 'docker' && (
        <GlassCard>
          <div className="flex items-center justify-between p-4 bg-blue-50 dark:bg-blue-950/30 rounded-lg border border-blue-200 dark:border-blue-800">
            <div className="flex items-center gap-3">
              <span className="text-3xl">🐳</span>
              <div>
                <h4 className="font-semibold text-blue-900 dark:text-blue-100">{t('clusters.dockerModeConnected')}</h4>
                <p className="text-sm text-blue-700 dark:text-blue-300">
                  {t('clusters.dockerModeDesc')}
                  {t('clusters.dockerModeHint')}
                </p>
              </div>
            </div>
            <div className="flex items-center gap-2">
              <span className="w-2 h-2 rounded-full bg-green-500 animate-pulse"></span>
              <span className="text-sm font-medium text-green-700 dark:text-green-300">{t('clusters.online')}</span>
            </div>
          </div>
        </GlassCard>
      )}

      {clusterInfo.k8s.connected && (
        <>
          <div className="flex gap-2 border-b border-gray-200 dark:border-gray-700 overflow-x-auto">
            {tabs.map(tab => (
              <button
                key={tab.id}
                onClick={() => setActiveTab(tab.id)}
                className={`px-4 py-2 font-medium text-sm whitespace-nowrap transition-colors ${
                  activeTab === tab.id
                    ? 'border-b-2 border-primary-600 text-primary-600'
                    : 'text-gray-500 hover:text-gray-700 dark:hover:text-gray-300'
                }`}
              >
                <span className="mr-2">{tab.icon}</span>
                {tab.label}
              </button>
            ))}
          </div>

          <div className="space-y-6">
            {activeTab === 'overview' && (
              <>
                <ChartCard title={t('clusters.clusterInfo')}>
                  <div className="grid grid-cols-2 md:grid-cols-4 gap-4">
                    <div className="p-4 bg-gray-50 dark:bg-gray-800/50 rounded-lg">
                      <p className="text-xs text-gray-500 mb-1">{t('clusters.version')}</p>
                      <p className="font-semibold text-gray-900 dark:text-white">{clusterInfo.k8s.version || '-'}</p>
                    </div>
                    <div className="p-4 bg-gray-50 dark:bg-gray-800/50 rounded-lg">
                      <p className="text-xs text-gray-500 mb-1">{t('clusters.namespace')}</p>
                      <p className="font-semibold text-gray-900 dark:text-white">{clusterInfo.k8s.namespace || '-'}</p>
                    </div>
                    <div className="p-4 bg-gray-50 dark:bg-gray-800/50 rounded-lg">
                      <p className="text-xs text-gray-500 mb-1">{t('clusters.inCluster')}</p>
                      <p className={`font-semibold ${
                        clusterInfo.k8s.inCluster ? 'text-green-600' : 'text-blue-600'
                      }`}>
                        {clusterInfo.k8s.inCluster ? t('clusters.yes') : t('clusters.no')}
                      </p>
                    </div>
                    <div className="p-4 bg-gray-50 dark:bg-gray-800/50 rounded-lg">
                      <p className="text-xs text-gray-500 mb-1">{t('clusters.nodes')}</p>
                      <p className="font-semibold text-gray-900 dark:text-white">{nodes.length}</p>
                    </div>
                  </div>
                </ChartCard>

                <GlassCard>
                  <h3 className="text-lg font-semibold mb-4">🔧 {t('clusters.platformServices')}</h3>
                  <div className="space-y-3">
                    {[
                      { name: 'api-gateway', desc: t('clusters.apiGatewayDesc'), port: 8080, icon: '🌐' },
                      { name: 'auth-service', desc: t('clusters.authServiceDesc'), port: 8082, icon: '🔐' },
                      { name: 'user-service', desc: t('clusters.userServiceDesc'), port: 8081, icon: '👤' },
                      { name: 'project-service', desc: t('clusters.projectServiceDesc'), port: 8083, icon: '📁' },
                      { name: 'generator-service', desc: t('clusters.generatorServiceDesc'), port: 8084, icon: '⚡' },
                      { name: 'operations-service', desc: t('clusters.operationsServiceDesc'), port: 8085, icon: '📊' },
                      { name: 'cluster-service', desc: t('clusters.clusterServiceDesc'), port: 8086, icon: '☸️' },
                      { name: 'postgres', desc: t('clusters.postgresDesc'), port: 5432, icon: '🐘' },
                      { name: 'redis', desc: t('clusters.redisDesc'), port: 6379, icon: '⚡' },
                      { name: 'web-admin', desc: t('clusters.webAdminDesc'), port: 3000, icon: '🖥️' },
                    ].map((service, index) => {
                      const pod = pods.find(p => p.name?.includes(service.name));
                      const isRunning = pod?.status === 'Running';
                      
                      return (
                        <div key={index} className="flex items-center justify-between p-4 rounded-xl bg-gray-50 dark:bg-gray-800/50 hover:bg-gray-100 dark:hover:bg-gray-800 transition-colors">
                          <div className="flex items-center gap-4">
                            <div className="w-10 h-10 rounded-lg bg-primary-100 dark:bg-primary-900/30 flex items-center justify-center text-xl">
                              {service.icon}
                            </div>
                            <div>
                              <p className="font-medium text-gray-900 dark:text-white">{service.name}</p>
                              <p className="text-sm text-gray-500">{service.desc}</p>
                            </div>
                          </div>
                          <div className="flex items-center gap-4">
                            <span className="px-2 py-1 text-xs font-mono bg-blue-100 text-blue-700 rounded">
                              :{service.port}
                            </span>
                            <span className={`w-2 h-2 rounded-full ${isRunning ? 'bg-green-500 animate-pulse' : 'bg-red-500'}`} />
                            <span className={`text-sm font-medium ${isRunning ? 'text-green-600' : 'text-red-600'}`}>
                              {isRunning ? t('clusters.running') : t('clusters.stopped')}
                            </span>
                          </div>
                        </div>
                      );
                    })}
                  </div>
                </GlassCard>

                <GlassCard>
                  <h3 className="text-lg font-semibold mb-4">{t('clusters.nodes')}</h3>
                  <div className="overflow-x-auto">
                    <table className="w-full">
                      <thead>
                        <tr className="border-b border-gray-200 dark:border-gray-700">
                          <th className="text-left py-3 px-4 text-sm font-medium text-gray-500">{t('clusters.node')}</th>
                          <th className="text-left py-3 px-4 text-sm font-medium text-gray-500">{t('common.status')}</th>
                          <th className="text-left py-3 px-4 text-sm font-medium text-gray-500">{t('clusters.roles')}</th>
                          <th className="text-left py-3 px-4 text-sm font-medium text-gray-500">{t('clusters.age')}</th>
                        </tr>
                      </thead>
                      <tbody>
                        {!Array.isArray(nodes) || nodes.length === 0 ? (
                          <tr><td colSpan="4" className="py-8 text-center text-gray-500">{t('clusters.noNodes')}</td></tr>
                        ) : nodes.map((node, index) => (
                          <tr key={index} className="border-b border-gray-100 dark:border-gray-800">
                            <td className="py-4 px-4 font-medium">{node.name}</td>
                            <td className="py-4 px-4">
                              <span className={`px-2 py-1 text-xs rounded-full ${
                                node.status === 'Ready' 
                                  ? 'bg-green-100 text-green-700' 
                                  : 'bg-yellow-100 text-yellow-700'
                              }`}>
                                {node.status}
                              </span>
                            </td>
                            <td className="py-4 px-4 text-gray-600">{node.roles?.join(', ') || '-'}</td>
                            <td className="py-4 px-4 text-gray-600">{node.age || '-'}</td>
                          </tr>
                        ))}
                      </tbody>
                    </table>
                  </div>
                </GlassCard>
              </>
            )}

            {activeTab === 'services' && (
              <GlassCard>
                <h3 className="text-lg font-semibold mb-4">{t('clusters.k8sServices')}</h3>
                <div className="overflow-x-auto">
                  <table className="w-full">
                    <thead>
                      <tr className="border-b border-gray-200 dark:border-gray-700">
                        <th className="text-left py-3 px-4 text-sm font-medium text-gray-500">{t('clusters.name')}</th>
                        <th className="text-left py-3 px-4 text-sm font-medium text-gray-500">{t('clusters.type')}</th>
                        <th className="text-left py-3 px-4 text-sm font-medium text-gray-500">{t('clusters.clusterIP')}</th>
                        <th className="text-left py-3 px-4 text-sm font-medium text-gray-500">{t('clusters.ports')}</th>
                      </tr>
                    </thead>
                    <tbody>
                      {!Array.isArray(services) || services.length === 0 ? (
                        <tr><td colSpan="4" className="py-8 text-center text-gray-500">{t('clusters.noServices')}</td></tr>
                      ) : services.map((service, index) => {
                        const serviceName = service.name || service.metadata?.name || `service-${index}`;
                        const serviceType = service.spec?.type || service.type || 'ClusterIP';
                        const clusterIP = service.spec?.cluster_ip || service.clusterIP || service.spec?.clusterIP || '-';
                        
                        let portsStr = '-';
                        if (Array.isArray(service.ports)) {
                          if (typeof service.ports[0] === 'object' && service.ports[0] !== null) {
                            portsStr = service.ports.map(p => p.port || p.target_port).filter(Boolean).join(', ');
                          } else if (typeof service.ports[0] === 'string' || typeof service.ports[0] === 'number') {
                            portsStr = service.ports.join(', ');
                          }
                        } else if (typeof service.ports === 'string' && service.ports) {
                          portsStr = service.ports;
                        } else if (Array.isArray(service.spec?.ports) && service.spec.ports.length > 0) {
                          portsStr = service.spec.ports.map(p => p.port).join(', ');
                        }
                        
                        return (
                          <tr key={index} className="border-b border-gray-100 dark:border-gray-800">
                            <td className="py-4 px-4 font-medium">{serviceName}</td>
                            <td className="py-4 px-4 text-gray-600">{serviceType}</td>
                            <td className="py-4 px-4 text-gray-600">{clusterIP}</td>
                            <td className="py-4 px-4 text-gray-600">{portsStr}</td>
                          </tr>
                        );
                      })}
                    </tbody>
                  </table>
                </div>
              </GlassCard>
            )}

            {activeTab === 'pods' && (
              <GlassCard>
                <h3 className="text-lg font-semibold mb-4">{t('clusters.podsIn').replace('{namespace}', selectedNamespace)}</h3>
                
                <div className="overflow-x-auto">
                  <table className="w-full">
                    <thead>
                      <tr className="border-b border-gray-200 dark:border-gray-700">
                        <th className="text-left py-3 px-4 text-sm font-medium text-gray-500">{t('clusters.name')}</th>
                        <th className="text-left py-3 px-4 text-sm font-medium text-gray-500">{t('common.status')}</th>
                        <th className="text-left py-3 px-4 text-sm font-medium text-gray-500">{t('clusters.ready')}</th>
                        <th className="text-right py-3 px-4 text-sm font-medium text-gray-500">{t('common.actions')}</th>
                      </tr>
                    </thead>
                    <tbody>
                      {!Array.isArray(pods) || pods.length === 0 ? (
                        <tr><td colSpan="4" className="py-8 text-center text-gray-500">{t('clusters.noPods')}</td></tr>
                      ) : pods.map((pod, index) => {
                        const podName = pod.name || pod.metadata?.name || `pod-${index}`;
                        const podStatus = pod.status || pod.status?.phase || 'Unknown';
                        const podReady = pod.ready || `${pod.status?.containerStatuses?.ready || 0}/${pod.status?.containerStatuses?.total || 0}`;
                        
                        return (
                          <tr key={index} className="border-b border-gray-100 dark:border-gray-800">
                            <td className="py-4 px-4 font-medium">{podName}</td>
                            <td className="py-4 px-4">
                              <span className={`px-2 py-1 text-xs rounded-full ${
                                podStatus === 'Running' 
                                  ? 'bg-green-100 text-green-700' 
                                  : podStatus === 'Pending'
                                  ? 'bg-yellow-100 text-yellow-700'
                                  : 'bg-red-100 text-red-700'
                              }`}>
                                {podStatus}
                              </span>
                            </td>
                            <td className="py-4 px-4 text-gray-600">{podReady}</td>
                            <td className="py-4 px-4 text-right">
                              <div className="flex items-center justify-end gap-2">
                                <button
                                  onClick={() => handleViewPodLogs({ ...pod, name: podName })}
                                  className="px-3 py-1 text-xs bg-blue-100 text-blue-700 rounded hover:bg-blue-200"
                                >
                                  📋 {t('clusters.logs')}
                                </button>
                              </div>
                            </td>
                          </tr>
                        );
                      })}
                    </tbody>
                  </table>
                </div>
              </GlassCard>
            )}

            {activeTab === 'deployments' && (
              <GlassCard>
                <h3 className="text-lg font-semibold mb-4">{t('clusters.deployments')}</h3>
                <div className="overflow-x-auto">
                  <table className="w-full">
                    <thead>
                      <tr className="border-b border-gray-200 dark:border-gray-700">
                        <th className="text-left py-3 px-4 text-sm font-medium text-gray-500">{t('clusters.name')}</th>
                        <th className="text-left py-3 px-4 text-sm font-medium text-gray-500">{t('clusters.replicas')}</th>
                        <th className="text-left py-3 px-4 text-sm font-medium text-gray-500">{t('common.status')}</th>
                        <th className="text-right py-3 px-4 text-sm font-medium text-gray-500">{t('common.actions')}</th>
                      </tr>
                    </thead>
                    <tbody>
                      {!Array.isArray(deployments) || deployments.length === 0 ? (
                        <tr><td colSpan="4" className="py-8 text-center text-gray-500">{t('clusters.noDeployments')}</td></tr>
                      ) : deployments.map((deployment, index) => (
                        <tr key={index} className="border-b border-gray-100 dark:border-gray-800">
                          <td className="py-4 px-4 font-medium">{deployment.name}</td>
                          <td className="py-4 px-4 text-gray-600">{deployment.ready_replicas}/{deployment.replicas}</td>
                          <td className="py-4 px-4">
                            {(() => {
                              const avail = deployment.available_replicas ?? 0;
                              const ready = deployment.ready_replicas ?? 0;
                              const total = deployment.replicas ?? 1;
                              const isReady = avail === total || ready === total;
                              return (
                                <span className={`px-2 py-1 text-xs rounded-full ${
                                  isReady
                                    ? 'bg-green-100 text-green-700'
                                    : 'bg-yellow-100 text-yellow-700'
                                }`}>
                                  {isReady ? 'Ready' : 'Updating'}
                                </span>
                              );
                            })()}
                          </td>
                          <td className="py-4 px-4 text-right">
                            <div className="flex items-center justify-end gap-2">
                              <button 
                                onClick={() => handleRestartDeployment(deployment)}
                                className="px-3 py-1 text-xs bg-blue-100 text-blue-700 rounded hover:bg-blue-200"
                              >
                                🔄 {t('clusters.restart')}
                              </button>
                              <button 
                                onClick={() => {
                                  setSelectedDeployment(deployment);
                                  setScaleReplicas(deployment.replicas);
                                  setShowScaleModal(true);
                                }}
                                className="px-3 py-1 text-xs bg-purple-100 text-purple-700 rounded hover:bg-purple-200"
                              >
                                ⚖️ {t('clusters.scale')}
                              </button>
                            </div>
                          </td>
                        </tr>
                      ))}
                    </tbody>
                  </table>
                </div>
              </GlassCard>
            )}

            {activeTab === 'health' && (
              <div className="space-y-6">
                <GlassCard>
                  <h3 className="text-lg font-semibold mb-4">✨ {t('clusters.healthCheck')} & {t('clusters.autoKeepalive')}</h3>

                  <div className="grid grid-cols-1 md:grid-cols-2 gap-6 mb-6">
                    <div className="p-4 rounded-xl bg-gray-50 dark:bg-gray-800/50">
                      <div className="flex items-center justify-between mb-3">
                        <span className="font-medium text-gray-700 dark:text-gray-300">{t('clusters.autoKeepalive')}</span>
                        <button
                          onClick={() => setAutoKeepalive(!autoKeepalive)}
                          className={`relative w-12 h-6 rounded-full transition-colors ${
                            autoKeepalive ? 'bg-green-500' : 'bg-gray-300'
                          }`}
                        >
                          <span className={`absolute top-0.5 left-0.5 w-5 h-5 bg-white rounded-full shadow transition-transform ${
                            autoKeepalive ? 'translate-x-6' : ''
                          }`} />
                        </button>
                      </div>
                      <p className="text-xs text-gray-500 dark:text-gray-400">
                        {autoKeepalive
                          ? t('clusters.keepaliveEnabled')
                          : t('clusters.keepaliveDisabled')
                        }
                      </p>
                    </div>

                    <div className="p-4 rounded-xl bg-gray-50 dark:bg-gray-800/50">
                      <label className="block font-medium text-gray-700 dark:text-gray-300 mb-2">
                        {t('clusters.checkInterval')} ({keepaliveInterval}s)
                      </label>
                      <input
                        type="range"
                        min="10"
                        max="120"
                        value={keepaliveInterval}
                        onChange={(e) => setKeepaliveInterval(parseInt(e.target.value))}
                        className="w-full"
                        disabled={!autoKeepalive}
                      />
                      <div className="flex justify-between text-xs text-gray-500 mt-1">
                        <span>10s</span>
                        <span>60s</span>
                        <span>120s</span>
                      </div>
                    </div>
                  </div>

                  <div className="flex gap-3 mb-6">
                    <button
                      onClick={runHealthCheck}
                      className="btn-primary flex items-center gap-2"
                    >
                      🔍 {t('clusters.runHealthCheck')}
                    </button>
                    {lastHealthCheck && (
                      <span className="text-sm text-gray-500 flex items-center">
                        {t('clusters.lastCheck')}: {lastHealthCheck}
                      </span>
                    )}
                  </div>

                  <div className="grid grid-cols-1 md:grid-cols-3 gap-4 mb-6">
                    <div className="p-4 rounded-lg bg-green-50 dark:bg-green-900/20 border border-green-200 dark:border-green-800">
                      <p className="text-sm text-green-600 dark:text-green-400 font-medium">✅ {t('clusters.healthyPods')}</p>
                      <p className="text-2xl font-bold text-green-700 dark:text-green-300 mt-1">
                        {healthStats.healthy} / {healthStats.total}
                      </p>
                    </div>
                    <div className="p-4 rounded-lg bg-yellow-50 dark:bg-yellow-900/20 border border-yellow-200 dark:border-yellow-800">
                      <p className="text-sm text-yellow-600 dark:text-yellow-400 font-medium">⚠️ {t('clusters.unhealthyPods')}</p>
                      <p className="text-2xl font-bold text-yellow-700 dark:text-yellow-300 mt-1">
                        {healthStats.unhealthy}
                      </p>
                    </div>
                    <div className="p-4 rounded-lg bg-blue-50 dark:bg-blue-900/20 border border-blue-200 dark:border-blue-800">
                      <p className="text-sm text-blue-600 dark:text-blue-400 font-medium">📊 {t('clusters.totalChecks')}</p>
                      <p className="text-2xl font-bold text-blue-700 dark:text-blue-300 mt-1">
                        {healthLog.length}
                      </p>
                    </div>
                  </div>
                </GlassCard>

                <GlassCard>
                  <h3 className="text-lg font-semibold mb-4">{t('clusters.healthLog')}</h3>
                  <div className="space-y-2 max-h-[400px] overflow-y-auto">
                    {!Array.isArray(healthLog) || healthLog.length === 0 ? (
                      <p className="text-gray-500 text-center py-8">{t('clusters.noHealthChecks')}</p>
                    ) : healthLog.map((log, index) => (
                      <div key={index} className={`p-3 rounded-lg ${
                        log.status === 'healthy'
                          ? 'bg-green-50 dark:bg-green-900/20 border-l-4 border-green-500'
                          : log.status === 'warning'
                          ? 'bg-yellow-50 dark:bg-yellow-900/20 border-l-4 border-yellow-500'
                          : 'bg-blue-50 dark:bg-blue-900/20 border-l-4 border-blue-500'
                      }`}>
                        <div className="flex items-start justify-between">
                          <p className="font-medium text-sm text-gray-800 dark:text-gray-200">{log.message}</p>
                          <span className="text-xs text-gray-500 ml-2 whitespace-nowrap">{log.time}</span>
                        </div>
                      </div>
                    ))}
                  </div>
                </GlassCard>
              </div>
            )}

            {activeTab === 'nodes' && (
              <div className="space-y-6">
                {/* 节点扩缩容操作面板 */}
                <GlassCard>
                  <div className="flex items-center justify-between mb-4">
                    <h3 className="text-lg font-semibold">{t('clusters.nodeScalingMgmt')}</h3>
                    <button onClick={() => { fetchNodeScalingInfo(); fetchNodesDetailed(); }} className="btn-secondary text-sm">
                      🔄 {t('common.refresh')}
                    </button>
                  </div>

                  {nodeScalingInfo?.supports_node_scaling ? (
                    <>
                      {/* 当前状态概览 */}
                      <div className="grid grid-cols-1 md:grid-cols-4 gap-4 mb-6">
                        <div className="p-4 rounded-lg bg-primary-50 dark:bg-primary-900/20 border border-primary-200 dark:border-primary-800">
                          <p className="text-xs text-primary-600 dark:text-primary-400 font-medium">{t('clusters.currentTotalNodes')}</p>
                          <p className="text-3xl font-bold text-primary-700 dark:text-primary-300">{nodeScalingInfo.current_node_count || nodes.length}</p>
                        </div>
                        <div className="p-4 rounded-lg bg-green-50 dark:bg-green-900/20 border border-green-200 dark:border-green-800">
                          <p className="text-xs text-green-600 dark:text-green-400 font-medium">{t('clusters.masterNodes')}</p>
                          <p className="text-3xl font-bold text-green-700 dark:text-green-300">{nodeScalingInfo.master_count || 1}</p>
                        </div>
                        <div className="p-4 rounded-lg bg-blue-50 dark:bg-blue-900/20 border border-blue-200 dark:border-blue-800">
                          <p className="text-xs text-blue-600 dark:text-blue-400 font-medium">{t('clusters.workerNodes')}</p>
                          <p className="text-3xl font-bold text-blue-700 dark:text-blue-300">{nodeScalingInfo.worker_count || (nodes.length - 1)}</p>
                        </div>
                        <div className="p-4 rounded-lg bg-purple-50 dark:bg-purple-900/20 border border-purple-200 dark:border-purple-800">
                          <p className="text-xs text-purple-600 dark:text-purple-400 font-medium">{t('clusters.envType')}</p>
                          <p className="text-lg font-bold text-purple-700 dark:text-purple-300">k3d ☸️</p>
                        </div>
                      </div>

                      {/* 操作按钮 */}
                      <div className="flex flex-wrap gap-4 mb-6">
                        <div className="flex items-center gap-2 text-sm text-gray-500">
                          <span className="w-2 h-2 rounded-full bg-green-500 animate-pulse"></span>
                          {t('clusters.k3dDetected')}
                        </div>
                      </div>
                    </>
                  ) : (
                    <div className="mb-6 p-4 rounded-lg border border-yellow-200 dark:border-yellow-800 bg-yellow-50 dark:bg-yellow-900/20">
                      <p className="text-sm text-yellow-800 dark:text-yellow-300">
                        {t('clusters.noScalingSupport')}
                      </p>
                    </div>
                  )}
                </GlassCard>

                {/* 节点列表（带操作按钮） */}
                <GlassCard>
                  <h3 className="text-lg font-semibold mb-4">{t('clusters.nodeList')}</h3>
                  <div className="overflow-x-auto">
                    <table className="w-full">
                      <thead>
                        <tr className="border-b border-gray-200 dark:border-gray-700">
                          <th className="text-left py-3 px-4 text-sm font-medium text-gray-500">{t('clusters.nodeName')}</th>
                          <th className="text-left py-3 px-4 text-sm font-medium text-gray-500">{t('common.status')}</th>
                          <th className="text-left py-3 px-4 text-sm font-medium text-gray-500">{t('common.role')}</th>
                          <th className="text-left py-3 px-4 text-sm font-medium text-gray-500">{t('clusters.podCount')}</th>
                          <th className="text-left py-3 px-4 text-sm font-medium text-gray-500">{t('common.age')}</th>
                          <th className="text-right py-3 px-4 text-sm font-medium text-gray-500">{t('common.actions')}</th>
                        </tr>
                      </thead>
                      <tbody>
                        {!Array.isArray(detailedNodes) || detailedNodes.length === 0 ? (
                          <tr><td colSpan="6" className="py-8 text-center text-gray-500">{t('clusters.noNodeData')}</td></tr>
                        ) : detailedNodes.map((node, index) => (
                          <tr key={index} className={`border-b border-gray-100 dark:border-gray-800 ${node.status !== 'Ready' ? 'bg-red-50 dark:bg-red-900/10' : ''}`}>
                            <td className="py-3 px-4">
                              <div className="font-mono text-sm font-medium">{node.name}</div>
                              {node.is_k3d_node && (
                                <span className="text-xs text-purple-500">k3d</span>
                              )}
                            </td>
                            <td className="py-3 px-4">
                              <span className={`inline-flex items-center px-2 py-1 rounded-full text-xs font-medium ${
                                node.status === 'Ready'
                                  ? 'bg-green-100 text-green-800 dark:bg-green-900/30 dark:text-green-400'
                                  : 'bg-yellow-100 text-yellow-800 dark:bg-yellow-900/30 dark:text-yellow-400'
                              }`}>
                                <span className={`w-1.5 h-1.5 rounded-full mr-1.5 ${node.status === 'Ready' ? 'bg-green-500' : 'bg-yellow-500'}`}></span>
                                {node.status || 'Unknown'}
                              </span>
                            </td>
                            <td className="py-3 px-4">
                              <span className={`text-xs px-2 py-0.5 rounded ${
                                node.roles?.includes('control-plane') || node.roles?.includes('master')
                                  ? 'bg-purple-100 text-purple-700'
                                  : 'bg-blue-100 text-blue-700'
                              }`}>
                                {Array.isArray(node.roles) && node.roles.length > 0
                                  ? node.roles.join(', ')
                                  : '-'}
                              </span>
                            </td>
                            <td className="py-3 px-4 text-sm">
                              <span className="px-2 py-0.5 bg-gray-100 dark:bg-gray-800 rounded text-xs">
                                {node.pod_count ?? 0}
                              </span>
                            </td>
                            <td className="py-3 px-4 text-sm text-gray-500">{node.age || '-'}</td>
                            <td className="py-3 px-4 text-right">
                              <div className="flex items-center justify-end gap-2">
                                <button
                                  onClick={() => handleRestartNode(node.name)}
                                  disabled={scalingLoading}
                                  className="px-3 py-1 text-xs bg-blue-100 text-blue-700 rounded hover:bg-blue-200 disabled:opacity-50"
                                  title={t('clusters.restartNodeTitle')}
                                >
                                  🔄 {t('clusters.restartBtn')}
                                </button>
                              </div>
                            </td>
                          </tr>
                        ))}
                      </tbody>
                    </table>
                  </div>
                </GlassCard>

                {/* 节点资源详情 */}
                {detailedNodes?.length > 0 && (
                  <GlassCard>
                    <h3 className="text-lg font-semibold mb-4">{t('clusters.nodeResourceDetails')}</h3>
                    <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-4">
                      {detailedNodes.map((node, index) => (
                        <div key={index} className="p-4 rounded-lg bg-gray-50 dark:bg-gray-800/50 border border-gray-200 dark:border-gray-700">
                          <div className="flex items-center justify-between mb-3">
                            <span className="font-mono text-sm font-semibold truncate">{node.name}</span>
                            <span className={`w-2 h-2 rounded-full ${node.status === 'Ready' ? 'bg-green-500' : 'bg-yellow-500'}`}></span>
                          </div>
                          <div className="space-y-2 text-xs">
                            <div className="flex justify-between">
                              <span className="text-gray-500">{t('clusters.cpuCapacity')}</span>
                              <span className="font-mono">{node.capacity?.cpu || '-'}</span>
                            </div>
                            <div className="flex justify-between">
                              <span className="text-gray-500">{t('clusters.cpuAvailable')}</span>
                              <span className="font-mono text-green-600">{node.allocatable?.cpu || '-'}</span>
                            </div>
                            <div className="flex justify-between">
                              <span className="text-gray-500">{t('clusters.memCapacity')}</span>
                              <span className="font-mono">{node.capacity?.memory || '-'}</span>
                            </div>
                            <div className="flex justify-between">
                              <span className="text-gray-500">{t('clusters.memAvailable')}</span>
                              <span className="font-mono text-green-600">{node.allocatable?.memory || '-'}</span>
                            </div>
                            <div className="flex justify-between">
                              <span className="text-gray-500">{t('clusters.podLimit')}</span>
                              <span className="font-mono">{node.capacity?.pods || '-'}</span>
                            </div>
                          </div>
                        </div>
                      ))}
                    </div>
                  </GlassCard>
                )}
              </div>
            )}
          </div>

          {/* 自动保活已整合到"健康检查"Tab中 */}
        </>
      )}

      {showLogModal && selectedPod && (
        <div className="fixed inset-0 bg-black bg-opacity-50 flex items-center justify-center p-4 z-50">
          <div className="bg-white dark:bg-gray-800 rounded-xl p-6 w-full max-w-4xl max-h-[80vh] overflow-hidden flex flex-col">
            <div className="flex items-center justify-between mb-4">
              <h3 className="text-lg font-semibold">📋 {t('clusters.podLogs')}: {selectedPod.name}</h3>
              <button 
                onClick={() => setShowLogModal(false)}
                className="text-gray-400 hover:text-gray-600"
              >✕</button>
            </div>
            <div className="bg-gray-900 rounded-lg p-4 flex-1 overflow-auto">
              <pre className="text-green-400 text-sm font-mono whitespace-pre-wrap">{podLogs}</pre>
            </div>
          </div>
        </div>
      )}

      {showScaleModal && selectedDeployment && (
        <div className="fixed inset-0 bg-black bg-opacity-50 flex items-center justify-center p-4 z-50">
          <div className="bg-white dark:bg-gray-800 rounded-xl p-6 w-full max-w-md">
            <h3 className="text-lg font-semibold mb-4">⚖️ {t('clusters.scaleDeployment')}: {selectedDeployment.name}</h3>
            
            <div className="space-y-4">
              <div>
                <label className="block text-sm font-medium text-gray-700 mb-2">
                  {t('clusters.replicaCount')}
                </label>
                <input 
                  type="number"
                  min="1"
                  max="10"
                  value={scaleReplicas}
                  onChange={(e) => setScaleReplicas(e.target.value)}
                  className="input-field w-full"
                />
              </div>
              
              <div className="flex gap-3">
                <button 
                  onClick={() => setShowScaleModal(false)}
                  className="flex-1 btn-secondary"
                >
                  {t('common.cancel')}
                </button>
                <button 
                  onClick={handleScaleDeployment}
                  className="flex-1 btn-primary"
                >
                  {t('clusters.applyScaling')}
                </button>
              </div>
            </div>
          </div>
        </div>
      )}
    </div>
  );
}
