import { useState, useEffect } from 'react';
import { useNavigate } from 'react-router-dom';
import { useTranslation } from 'react-i18next';
import { projectAPI, operationsAPI, generatorAPI } from '../api';
import { GlassCard, StatCard } from '../components/Cards';

export default function Projects() {
  const navigate = useNavigate();
  const [projects, setProjects] = useState([]);
  const [loading, setLoading] = useState(true);
  const [showModal, setShowModal] = useState(false);
  const [newProject, setNewProject] = useState({ name: '', description: '' });
  const [showEditModal, setShowEditModal] = useState(false);
  const [editProject, setEditProject] = useState(null);
  const [showDeleteModal, setShowDeleteModal] = useState(false);
  const [deleteId, setDeleteId] = useState(null);
  const { t } = useTranslation();

  useEffect(() => {
    fetchProjects();
  }, []);

  const fetchProjects = async () => {
    setLoading(true);
    try {
      const response = await projectAPI.list();
      setProjects(response.data?.data || response.data || []);
    } catch (error) {
      console.error('Failed to fetch projects:', error);
    } finally {
      setLoading(false);
    }
  };

  const handleCreate = async () => {
    if (!newProject.name.trim()) return;
    try {
      await projectAPI.create(newProject);
      setShowModal(false);
      setNewProject({ name: '', description: '' });
      fetchProjects();
    } catch (error) {
      console.error('Failed to create project:', error);
    }
  };

  const handleUpdate = async () => {
    if (!editProject.name.trim()) return;
    try {
      await projectAPI.update(editProject.id, editProject);
      setShowEditModal(false);
      setEditProject(null);
      fetchProjects();
    } catch (error) {
      console.error('Failed to update project:', error);
    }
  };

  const openEditModal = (project) => {
    setEditProject(project);
    setShowEditModal(true);
  };

  const confirmDelete = (id) => {
    setDeleteId(id);
    setShowDeleteModal(true);
  };

  const handleDelete = async () => {
    if (!deleteId) return;
    try {
      await projectAPI.delete(deleteId);
      setShowDeleteModal(false);
      setDeleteId(null);
      fetchProjects();
    } catch (error) {
      console.error('Failed to delete project:', error);
    }
  };

  const handleRegenerate = async (project) => {
    if (!window.confirm(t('projects.regenerateConfirm') + ` "${project.name}"?`)) return;
    try {
      const startTime = Date.now();
      await projectAPI.regenerate(project.id);

      operationsAPI.recordOperationLog({
        action: 'regenerate',
        resource: 'project',
        resource_id: project.id,
        details: JSON.stringify({ project_name: project.name }),
        status: 'success',
        duration: Date.now() - startTime,
      }).catch(() => {});

      alert(t('projects.regenerateSuccess'));
      fetchProjects();
    } catch (error) {
      const msg = error.response?.data?.message || error.message;

      operationsAPI.recordOperationLog({
        action: 'regenerate',
        resource: 'project',
        resource_id: project.id,
        details: JSON.stringify({ project_name: project.name, error: msg }),
        status: 'failed',
        duration: Date.now() - Date.now(),
        error: msg,
      }).catch(() => {});

      alert(`${t('projects.regenerateFailed')}: ${msg}`);
    }
  };

  const handleDownload = async (project) => {
    if (project.status !== 'generated') {
      alert(t('projects.downloadNotReady'));
      return;
    }

    try {
      const startTime = Date.now();

      const response = await generatorAPI.download(project.id);

      const blob = new Blob([response.data], { type: 'application/zip' });
      const url = window.URL.createObjectURL(blob);
      const link = document.createElement('a');
      link.href = url;
      link.download = `${project.name || 'project'}_${new Date().toISOString().slice(0,10)}.zip`;
      document.body.appendChild(link);
      link.click();
      document.body.removeChild(link);
      window.URL.revokeObjectURL(url);

      operationsAPI.recordOperationLog({
        action: 'download',
        resource: 'project',
        resource_id: project.id,
        details: JSON.stringify({ project_name: project.name }),
        status: 'success',
        duration: Date.now() - startTime,
      }).catch(() => {});
    } catch (error) {
      const msg = error.response?.data?.message || error.message;

      operationsAPI.recordOperationLog({
        action: 'download',
        resource: 'project',
        resource_id: project.id,
        details: JSON.stringify({ project_name: project.name, error: msg }),
        status: 'failed',
        duration: Date.now() - Date.now(),
        error: msg,
      }).catch(() => {});

      alert(`${t('projects.downloadFailed')}: ${msg}`);
    }
  };

  return (
    <div className="space-y-6">
      <div className="flex items-center justify-between">
        <div>
          <h1 className="text-2xl font-bold text-gray-900 dark:text-white">{t('projects.title')}</h1>
          <p className="text-gray-500 dark:text-gray-400 mt-1">{t('projects.subtitle')}</p>
        </div>
        <button
          onClick={() => navigate('/generator')}
          className="btn-primary flex items-center gap-2"
        >
          <svg className="w-5 h-5" fill="none" viewBox="0 0 24 24" stroke="currentColor">
            <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M12 4v16m8-8H4" />
          </svg>
          {t('projects.newProject')}
        </button>
      </div>

      <div className="grid grid-cols-1 md:grid-cols-3 gap-6 mb-6">
        <StatCard title={t('projects.totalProjects')} value={projects.length} />
        <StatCard title={t('projects.generated')} value={projects.filter(p => p.status === 'generated').length} />
        <StatCard title={t('projects.pending')} value={projects.filter(p => p.status === 'pending').length} />
      </div>

      <GlassCard>
        <div className="overflow-x-auto">
          <table className="min-w-full divide-y divide-gray-200 dark:divide-gray-700">
            <thead>
              <tr>
                <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 dark:text-gray-400 uppercase tracking-wider">{t('projects.projectName')}</th>
                <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 dark:text-gray-400 uppercase tracking-wider">{t('common.description')}</th>
                <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 dark:text-gray-400 uppercase tracking-wider">{t('common.status')}</th>
                <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 dark:text-gray-400 uppercase tracking-wider">{t('common.createdAt')}</th>
                <th className="px-6 py-3 text-right text-xs font-medium text-gray-500 dark:text-gray-400 uppercase tracking-wider">{t('common.actions')}</th>
              </tr>
            </thead>
            <tbody className="bg-white dark:bg-gray-900 divide-y divide-gray-200 dark:divide-gray-800">
              {loading ? (
                <tr><td colSpan={5} className="px-6 py-12 text-center text-gray-400"><svg className="w-8 h-8 animate-spin mx-auto text-primary-500" fill="none" viewBox="0 0 24 24"><circle className="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" strokeWidth="4"/><path className="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4z"/></svg></td></tr>
              ) : projects.length === 0 ? (
                <tr><td colSpan={5} className="px-6 py-12 text-center text-gray-400">
                  <p>{t('projects.noProjects')}</p>
                  <button onClick={() => navigate('/generator')} className="mt-3 btn-secondary text-sm px-4 py-2">{t('projects.createFirstProject')}</button>
                </td></tr>
              ) : projects.map((project) => (
                <tr key={project.id} className="hover:bg-gray-50 dark:hover:bg-gray-800/50 transition-colors">
                  <td className="px-6 py-4 whitespace-nowrap">
                    <div className="flex items-center gap-3">
                      <div className={`w-10 h-10 rounded-lg flex items-center justify-center ${project.status === 'generated' ? 'bg-green-100 dark:bg-green-900/30' : 'bg-yellow-100 dark:bg-yellow-900/30'}`}>
                        <svg className={`w-5 h-5 ${project.status === 'generated' ? 'text-green-600' : 'text-yellow-600'}`} fill="none" viewBox="0 0 24 24" stroke="currentColor"><path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M9 12h6m-6 4h6m2 5H7a2 2 0 01-2-2V5a2 2 0 012-2h5.586a1 1 0 01.707.293l5.414 5.414a1 1 0 01.293.707V19a2 2 0 01-2 2z"/></svg>
                      </div>
                      <span className="font-medium text-gray-900 dark:text-white">{project.name}</span>
                    </div>
                  </td>
                  <td className="px-6 py-4 max-w-[200px]"><span className="text-sm text-gray-500 dark:text-gray-400 truncate block">{project.description || '-'}</span></td>
                  <td className="px-6 py-4 whitespace-nowrap">
                    <span className={`inline-flex items-center px-2.5 py-0.5 rounded-full text-xs font-medium ${project.status === 'generated' ? 'bg-green-100 text-green-800 dark:bg-green-900/30 dark:text-green-400' : 'bg-yellow-100 text-yellow-800 dark:bg-yellow-900/30 dark:text-yellow-400'}`}>
                      {project.status === 'generated' ? t('common.generated') : t('common.pending')}
                    </span>
                  </td>
                  <td className="px-6 py-4 whitespace-nowrap text-sm text-gray-500 dark:text-gray-400">{new Date(project.created_at).toLocaleDateString()}</td>
                  <td className="px-6 py-4 whitespace-nowrap text-right">
                    <div className="flex items-center justify-end gap-2">
                      <button onClick={() => navigate(`/generator?projectId=${project.id}`)} className="p-2 hover:bg-blue-100 dark:hover:bg-blue-900/30 rounded-lg transition-colors" title={t('common.edit')}>
                        <svg className="w-5 h-5 text-blue-500" fill="none" viewBox="0 0 24 24" stroke="currentColor"><path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M15.232 5.232l3.536 3.536m-2.036-5.036a2.5 2.5 0 113.536 3.536L6.5 21.036H3v-3.572L16.732 3.732z"/></svg>
                      </button>
                      <button onClick={() => handleDownload(project)} className={`p-2 rounded-lg transition-colors ${project.status === 'generated' ? 'hover:bg-green-100 dark:hover:bg-green-900/30' : 'opacity-40 cursor-not-allowed'}`} title={project.status === 'generated' ? t('projects.download') : t('projects.downloadNotReady')}>
                        <svg className={`w-5 h-5 ${project.status === 'generated' ? 'text-green-500' : 'text-gray-400'}`} fill="none" viewBox="0 0 24 24" stroke="currentColor"><path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M4 16v1a3 3 0 003 3h10a3 3 0 003-3v-1m-4-4l-4 4m0 0l-4-4m4 4V4"/></svg>
                      </button>
                      <button onClick={() => handleRegenerate(project)} className="p-2 hover:bg-purple-100 dark:hover:bg-purple-900/30 rounded-lg transition-colors" title={t('projects.regenerate')}>
                        <svg className="w-5 h-5 text-purple-500" fill="none" viewBox="0 0 24 24" stroke="currentColor"><path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M4 4v5h.582m15.356 2A8.001 8.001 0 004.582 9m0 0H9m11 11v-5h-.581m0 0a8.003 8.003 0 01-15.357-2m15.357 2H15"/></svg>
                      </button>
                      <button onClick={() => confirmDelete(project.id)} className="p-2 hover:bg-red-100 dark:hover:bg-red-900/30 rounded-lg transition-colors" title={t('common.delete')}>
                        <svg className="w-5 h-5 text-red-500" fill="none" viewBox="0 0 24 24" stroke="currentColor"><path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M19 7l-.867 12.142A2 2 0 0116.138 21H7.862a2 2 0 01-1.995-1.858L5 7m5 4v6m4-6v6m1-10V4a1 1 0 00-1-1h-4a1 1 0 00-1 1v3M4 7h16"/></svg>
                      </button>
                    </div>
                  </td>
                </tr>
              ))}
            </tbody>
          </table>
        </div>
      </GlassCard>

      {showModal && (
        <div className="fixed inset-0 bg-black bg-opacity-50 flex items-center justify-center z-50 p-4" onClick={(e) => e.target === e.currentTarget && setShowModal(false)}>
          <div className="bg-white dark:bg-gray-800 rounded-xl shadow-xl w-full max-w-md p-6 space-y-4">
            <h3 className="text-lg font-semibold text-gray-900 dark:text-white">{t('projects.createProject')}</h3>
            <input type="text" value={newProject.name} onChange={(e) => setNewProject({...newProject, name: e.target.value})} className="input-field" placeholder={t('projects.projectName')}/>
            <textarea value={newProject.description} onChange={(e) => setNewProject({...newProject, description: e.target.value})} className="input-field" placeholder={t('projects.projectDescription')} rows={3}/>
            <div className="flex gap-3 pt-2">
              <button onClick={() => setShowModal(false)} className="flex-1 btn-secondary">{t('common.cancel')}</button>
              <button onClick={handleCreate} disabled={!newProject.name.trim()} className="flex-1 btn-primary disabled:opacity-50">{t('common.create')}</button>
            </div>
          </div>
        </div>
      )}

      {showEditModal && editProject && (
        <div className="fixed inset-0 bg-black bg-opacity-50 flex items-center justify-center z-50 p-4" onClick={(e) => e.target === e.currentTarget && setShowEditModal(false)}>
          <div className="bg-white dark:bg-gray-800 rounded-xl shadow-xl w-full max-w-md p-6 space-y-4">
            <h3 className="text-lg font-semibold text-gray-900 dark:text-white">{t('projects.editProject')}</h3>
            <div className="space-y-4">
              <div>
                <label className="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-2">{t('projects.projectName')}</label>
                <input type="text" value={editProject.name} onChange={(e) => setEditProject({...editProject, name: e.target.value})} className="input-field" placeholder={t('projects.projectName')}/>
              </div>
              <div>
                <label className="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-2">{t('common.description')}</label>
                <textarea value={editProject.description} onChange={(e) => setEditProject({...editProject, description: e.target.value})} className="input-field" placeholder={t('projects.projectDescription')} rows={3}/>
              </div>
            </div>
            <div className="flex gap-3 pt-2">
              <button onClick={() => {setShowEditModal(false); setEditProject(null)}} className="flex-1 btn-secondary">{t('common.cancel')}</button>
              <button onClick={handleUpdate} disabled={!editProject.name.trim()} className="flex-1 btn-primary disabled:opacity-50">{t('common.save')}</button>
            </div>
          </div>
        </div>
      )}

      {showDeleteModal && (
        <div className="fixed inset-0 bg-black bg-opacity-50 flex items-center justify-center z-50 p-4" onClick={(e) => e.target === e.currentTarget && setShowDeleteModal(false)}>
          <div className="bg-white dark:bg-gray-800 rounded-xl shadow-xl w-full max-w-sm p-6 space-y-4">
            <h3 className="text-lg font-semibold text-red-600">{t('projects.deleteConfirmTitle')}</h3>
            <p className="text-gray-600 dark:text-gray-400 text-sm">{t('projects.deleteConfirmMessage')}</p>
            <div className="flex gap-3 pt-2">
              <button onClick={() => {setShowDeleteModal(false); setDeleteId(null)}} className="flex-1 btn-secondary">{t('common.cancel')}</button>
              <button onClick={handleDelete} className="flex-1 bg-red-600 text-white px-4 py-2 rounded-lg hover:bg-red-700 transition-colors">{t('common.delete')}</button>
            </div>
          </div>
        </div>
      )}
    </div>
  );
}
