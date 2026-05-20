import axios from 'axios';

const API_BASE_URL = import.meta.env.VITE_API_URL || '/api/v1';
const CLUSTER_API_URL = import.meta.env.VITE_CLUSTER_API_URL || '/api/v1';

const api = axios.create({
  baseURL: API_BASE_URL,
  headers: {
    'Content-Type': 'application/json',
  },
  timeout: 60000,
});

// Cluster API 实例（可以单独配置）
const clusterApi = axios.create({
  baseURL: CLUSTER_API_URL,
  headers: {
    'Content-Type': 'application/json',
  },
  timeout: 60000,
});

clusterApi.interceptors.request.use((config) => {
  const token = localStorage.getItem('token');
  if (token) {
    config.headers.Authorization = `Bearer ${token}`;
  }
  return config;
});

clusterApi.interceptors.response.use(
  (response) => response,
  (error) => {
    if (error.response?.status === 401) {
      localStorage.removeItem('token');
      localStorage.removeItem('user');
      window.location.href = '/login';
    } else if (error.response?.status === 403) {
      console.warn('Access denied: insufficient permissions');
    } else if (!error.response && error.message === 'Network Error') {
      console.error('Network error - backend may be unavailable');
    }
    return Promise.reject(error);
  }
);

api.interceptors.request.use((config) => {
  const token = localStorage.getItem('token');
  if (token) {
    config.headers.Authorization = `Bearer ${token}`;
  }
  return config;
});

api.interceptors.response.use(
  (response) => response,
  (error) => {
    if (error.response?.status === 401) {
      localStorage.removeItem('token');
      localStorage.removeItem('user');
      window.location.href = '/login';
    } else if (error.response?.status === 403) {
      console.warn('Access denied: insufficient permissions');
    } else if (!error.response && error.message === 'Network Error') {
      console.error('Network error - backend may be unavailable');
    }
    return Promise.reject(error);
  }
);

export const authAPI = {
  login: (data) => api.post('/auth/login', data),
  register: (data) => api.post('/auth/register', data),
  getMe: () => api.get('/auth/me'),
};

export const projectAPI = {
  list: () => api.get('/projects'),
  get: (id) => api.get(`/projects/${id}`),
  create: (data) => api.post('/projects', data),
  update: (id, data) => api.put(`/projects/${id}`, data),
  delete: (id) => api.delete(`/projects/${id}`),
  regenerate: (id) => api.post(`/generator/generate/${id}`),
};

export const generatorAPI = {
  generate: (data) => api.post('/generator/generate', data),
  generateFromProject: (id) => api.post(`/generator/generate/${id}`),
  download: (projectId) => api.get(`/generator/download/${projectId}`, { responseType: 'blob' }),
  preview: (projectId) => api.get(`/generator/preview/${projectId}`),
};

export const userAPI = {
  list: () => api.get('/users'),
  get: (id) => api.get(`/users/${id}`),
  create: (data) => api.post('/users', data),
  update: (id, data) => api.put(`/users/${id}`, data),
  delete: (id) => api.delete(`/users/${id}`),
};

export const clusterAPI = {
  getStatus: () => clusterApi.get('/clusters/status'),
  getMetrics: () => clusterApi.get('/clusters/metrics'),
  getHealth: () => clusterApi.get('/clusters/health'),
  getK8sStatus: () => clusterApi.get('/clusters/k8s/status'),
  getK8sInfo: () => clusterApi.get('/clusters/k8s/info'),
  getK8sNamespaces: () => clusterApi.get('/clusters/k8s/namespaces'),
  getK8sNodes: () => clusterApi.get('/clusters/k8s/nodes'),
  getK8sPods: (namespace) => clusterApi.get('/clusters/k8s/pods', { params: { namespace } }),
  getK8sServices: (namespace) => clusterApi.get('/clusters/k8s/services', { params: { namespace } }),
  getK8sDeployments: (namespace) => clusterApi.get('/clusters/k8s/deployments', { params: { namespace } }),
  getK8sEvents: (namespace) => clusterApi.get('/clusters/k8s/events', { params: { namespace } }),
  getK8sPodLogs: (namespace, name) => clusterApi.get(`/clusters/k8s/pods/${namespace}/${name}/logs`),
  deleteK8sPod: (namespace, name) => clusterApi.delete(`/clusters/k8s/pods/${namespace}/${name}`),
  restartK8sDeployment: (namespace, name) => clusterApi.post(`/clusters/k8s/deployments/${namespace}/${name}/restart`),
  scaleK8sDeployment: (namespace, name, replicas) => clusterApi.post(`/clusters/k8s/deployments/${namespace}/${name}/scale`, { replicas }),
  getDockerServices: () => clusterApi.get('/clusters/docker/services'),

  // ===== 节点扩缩容API（k3d环境）=====
  getNodeScalingInfo: () => clusterApi.get('/clusters/nodes/scaling-info'),
  getNodesDetailed: () => clusterApi.get('/clusters/nodes/list-detailed'),
  scaleNode: (data) => clusterApi.post('/clusters/nodes/scale', data),
  checkK3dAvailable: () => clusterApi.get('/clusters/nodes/k3d-check'),

  // ===== 自动保活API =====
  getAutoHealingStatus: () => clusterApi.get('/clusters/auto-healing/status'),
  updateAutoHealingConfig: (config) => clusterApi.put('/clusters/auto-healing/config', config),
  getHealingHistory: () => clusterApi.get('/clusters/auto-healing/history'),
  triggerManualHealthCheck: () => clusterApi.post('/clusters/auto-healing/trigger'),
};

export const operationsAPI = {
  health: () => api.get('/operations/health'),
  stats: () => api.get('/operations/stats'),
  getMetrics: () => api.get('/operations/metrics'),
  getServices: () => api.get('/operations/services'),
  getEvents: (lang) => api.get(`/operations/events${lang ? '?lang=' + lang : ''}`),
  getOperationLogs: (params = {}) => api.get('/operations/operation-logs', { params }),
  recordOperationLog: (data) => api.post('/operations/operation-logs/record', data),
};

export default api;