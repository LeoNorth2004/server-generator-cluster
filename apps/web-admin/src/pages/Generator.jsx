import { useState, useEffect } from 'react';
import { useSearchParams, useNavigate } from 'react-router-dom';
import { useTranslation } from 'react-i18next';
import { generatorAPI, projectAPI, operationsAPI } from '../api';
import { GlassCard, StatCard } from '../components/Cards';

export default function Generator() {
  const [searchParams] = useSearchParams();
  const navigate = useNavigate();
  const editProjectId = searchParams.get('projectId');

  const [config, setConfig] = useState({
    project_name: '',
    tables: [{ name: '', fields: [{ name: '', type: 'varchar', comment: '' }] }],
  });
  const [generating, setGenerating] = useState(false);
  const [result, setResult] = useState(null);
  const [loadingProject, setLoadingProject] = useState(!!editProjectId);
  const { t, i18n } = useTranslation();

  useEffect(() => {
    if (editProjectId) {
      loadProjectForEdit(editProjectId);
    }
  }, [editProjectId]);

  const loadProjectForEdit = async (projectId) => {
    setLoadingProject(true);
    try {
      const response = await projectAPI.get(projectId);
      const project = response.data?.data || response.data;
      if (project && project.table_config) {
        let tables = [];
        try {
          tables = JSON.parse(project.table_config);
        } catch (e) {
          console.warn('Failed to parse table_config:', e);
        }
        if (Array.isArray(tables) && tables.length > 0) {
          setConfig({
            project_name: project.name || '',
            tables: tables,
          });
        }
      }
    } catch (error) {
      console.error('Failed to load project:', error);
      alert(t('generator.loadProjectFailed') || '加载项目失败');
    } finally {
      setLoadingProject(false);
    }
  };

  const addTable = () => {
    setConfig({
      ...config,
      tables: [...config.tables, { name: '', fields: [{ name: '', type: 'varchar', comment: '' }] }],
    });
  };

  const addField = (tableIndex) => {
    const newTables = [...config.tables];
    newTables[tableIndex].fields.push({ name: '', type: 'varchar', comment: '' });
    setConfig({ ...config, tables: newTables });
  };

  const removeTable = (tableIndex) => {
    if (config.tables.length <= 1) return;
    const newTables = config.tables.filter((_, i) => i !== tableIndex);
    setConfig({ ...config, tables: newTables });
  };

  const removeField = (tableIndex, fieldIndex) => {
    const newTables = [...config.tables];
    if (newTables[tableIndex].fields.length <= 1) return;
    newTables[tableIndex].fields = newTables[tableIndex].fields.filter((_, i) => i !== fieldIndex);
    setConfig({ ...config, tables: newTables });
  };

  const handleGenerate = async () => {
    if (!config.project_name) return;

    setGenerating(true);
    setResult(null);
    const startTime = Date.now();
    try {
      const response = await generatorAPI.generate(config);
      const data = response.data?.data || response.data;
      setResult(data);

      const duration = Date.now() - startTime;

      if (data?.project_id) {
        operationsAPI.recordOperationLog({
          action: editProjectId ? 'regenerate' : 'generate',
          resource: 'project',
          resource_id: data.project_id,
          details: JSON.stringify({ project_name: config.project_name, tables_count: config.tables.length, files_generated: data.code?.files ? Object.keys(data.code.files).length : 0 }),
          status: 'success',
          duration: duration,
        }).catch(() => {});
      }

      alert(t('generator.success') || '代码生成成功！');
    } catch (error) {
      console.error('Generation failed:', error);
      const msg = error.response?.data?.message || error.message || 'Unknown error';

      operationsAPI.recordOperationLog({
        action: 'generate',
        resource: 'project',
        resource_id: 0,
        details: JSON.stringify({ project_name: config.project_name, error: msg }),
        status: 'failed',
        duration: Date.now() - startTime,
        error: msg,
      }).catch(() => {});

      alert(`${t('generator.failed') || '代码生成失败'}: ${msg}`);
    } finally {
      setGenerating(false);
    }
  };

  const handleDownload = async () => {
    if (!result?.project_id) return;
    try {
      const response = await generatorAPI.download(result.project_id);
      if (response.data instanceof Blob && response.data.type === 'application/zip') {
        const url = URL.createObjectURL(response.data);
        const a = document.createElement('a');
        a.href = url;
        a.download = `${config.project_name}-generated-code.zip`;
        document.body.appendChild(a);
        a.click();
        document.body.removeChild(a);
        URL.revokeObjectURL(url);

        operationsAPI.recordOperationLog({
          action: 'download',
          resource: 'project',
          resource_id: result.project_id,
          details: JSON.stringify({ project_name: config.project_name }),
          status: 'success',
          duration: 0,
        }).catch(() => {});

        alert(t('generator.downloadSuccess') || '下载成功！');
      } else {
        const errorText = await response.data.text();
        alert(`下载失败: ${errorText}`);
      }
    } catch (error) {
      console.error('Download failed:', error);
      alert(t('generator.downloadFailed') || '下载失败');
    }
  };

  const handleRegenerate = async () => {
    if (!editProjectId) return;
    if (!window.confirm(t('projects.regenerateConfirm') + ` "${config.project_name}"?`)) return;
    setGenerating(true);
    setResult(null);
    const startTime = Date.now();
    try {
      const response = await generatorAPI.generateFromProject(editProjectId);
      const data = response.data?.data || response.data;
      setResult(data);

      operationsAPI.recordOperationLog({
        action: 'regenerate',
        resource: 'project',
        resource_id: parseInt(editProjectId),
        details: JSON.stringify({ project_name: config.project_name }),
        status: 'success',
        duration: Date.now() - startTime,
      }).catch(() => {});

      alert(t('projects.regenerateSuccess'));
    } catch (error) {
      const msg = error.response?.data?.message || error.message;
      operationsAPI.recordOperationLog({
        action: 'regenerate',
        resource: 'project',
        resource_id: parseInt(editProjectId),
        details: JSON.stringify({ project_name: config.project_name, error: msg }),
        status: 'failed',
        duration: Date.now() - startTime,
        error: msg,
      }).catch(() => {});
      alert(`${t('projects.regenerateFailed')}: ${msg}`);
    } finally {
      setGenerating(false);
    }
  };

  if (loadingProject) {
    return (
      <div className="flex items-center justify-center py-20">
        <svg className="w-8 h-8 animate-spin text-primary-500" fill="none" viewBox="0 0 24 24">
          <circle className="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" strokeWidth="4" />
          <path className="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4z" />
        </svg>
        <span className="ml-3 text-gray-600 dark:text-gray-400">{t('common.loading')}</span>
      </div>
    );
  }

  return (
    <div className="space-y-6">
      <div className="flex items-center justify-between">
        <div>
          <h1 className="text-2xl font-bold text-gray-900 dark:text-white">{t('generator.title')}</h1>
          <p className="text-gray-500 dark:text-gray-400 mt-1">
            {editProjectId
              ? `${t('generator.editingProject') || '编辑项目'}: ${config.project_name}`
              : t('generator.subtitle')}
          </p>
        </div>
        {editProjectId && (
          <button
            onClick={() => navigate('/projects')}
            className="btn-secondary flex items-center gap-2"
          >
            <svg className="w-4 h-4" fill="none" viewBox="0 0 24 24" stroke="currentColor">
              <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M15 19l-7-7 7-7" />
            </svg>
            {t('common.backToList') || '返回项目列表'}
          </button>
        )}
      </div>

      <div className="grid grid-cols-1 lg:grid-cols-3 gap-6">
        <div className="lg:col-span-2 space-y-6">
          <GlassCard>
            <h3 className="text-lg font-semibold text-gray-900 dark:text-white mb-4">{t('generator.projectConfig')}</h3>
            <div className="space-y-4">
              <div>
                <label className="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-2">{t('generator.projectName')}</label>
                <input
                  type="text"
                  value={config.project_name}
                  onChange={(e) => setConfig({ ...config, project_name: e.target.value })}
                  className="input-field"
                  placeholder="my-service"
                />
              </div>
            </div>
          </GlassCard>

          <GlassCard>
            <div className="flex items-center justify-between mb-4">
              <h3 className="text-lg font-semibold text-gray-900 dark:text-white">{t('generator.databaseTables')}</h3>
              <button onClick={addTable} className="btn-secondary text-sm px-4 py-2">{t('generator.addTable')}</button>
            </div>
            <div className="space-y-6">
              {config.tables.map((table, tableIndex) => (
                <div key={tableIndex} className="p-6 bg-gray-50 dark:bg-gray-800/50 rounded-xl border border-gray-200 dark:border-gray-700">
                  <div className="flex gap-4 mb-4 items-center">
                    <label className="text-sm font-medium text-gray-700 dark:text-gray-300 min-w-[80px]">表名:</label>
                    <input
                      type="text"
                      value={table.name}
                      onChange={(e) => {
                        const newTables = [...config.tables];
                        newTables[tableIndex].name = e.target.value;
                        setConfig({ ...config, tables: newTables });
                      }}
                      className="input-field flex-1 min-w-[200px] h-10 px-3 text-base"
                      placeholder="例如: users"
                    />
                    <button onClick={() => addField(tableIndex)} className="btn-secondary text-sm px-4 py-2 whitespace-nowrap">+ 添加字段</button>
                    {config.tables.length > 1 && (
                      <button onClick={() => removeTable(tableIndex)} className="p-2 hover:bg-red-100 dark:hover:bg-red-900/30 rounded-lg transition-colors" title="删除表">
                        <svg className="w-5 h-5 text-red-500" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                          <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M19 7l-.867 12.142A2 2 0 0116.138 21H7.862a2 2 0 01-1.995-1.858L5 7m5 4v6m4-6v6m1-10V4a1 1 0 00-1-1h-4a1 1 0 00-1 1v3M4 7h16" />
                        </svg>
                      </button>
                    )}
                  </div>
                  <div className="space-y-3 ml-[96px]">
                    <div className="text-sm font-medium text-gray-600 dark:text-gray-400 mb-2">字段列表:</div>
                    {table.fields.map((field, fieldIndex) => (
                      <div key={`${tableIndex}-${fieldIndex}`} className="flex gap-3 items-center p-3 bg-white dark:bg-gray-900 rounded-lg border border-gray-200 dark:border-gray-700">
                        <input
                          type="text"
                          value={field.name}
                          onChange={(e) => {
                            const newTables = [...config.tables];
                            newTables[tableIndex].fields[fieldIndex].name = e.target.value;
                            setConfig({ ...config, tables: newTables });
                          }}
                          className="input-field flex-1 min-w-[150px] h-9 px-3 text-sm"
                          placeholder="字段名称 (如: username)"
                        />
                        <select
                          value={field.type}
                          onChange={(e) => {
                            const newTables = [...config.tables];
                            newTables[tableIndex].fields[fieldIndex].type = e.target.value;
                            setConfig({ ...config, tables: newTables });
                          }}
                          className="input-field w-36 h-9 px-2 text-sm cursor-pointer"
                        >
                          <option value="varchar">varchar</option>
                          <option value="bigint">bigint</option>
                          <option value="int">int</option>
                          <option value="decimal">decimal</option>
                          <option value="timestamp">timestamp</option>
                          <option value="boolean">boolean</option>
                          <option value="text">text</option>
                        </select>
                        <input
                          type="text"
                          value={field.comment}
                          onChange={(e) => {
                            const newTables = [...config.tables];
                            newTables[tableIndex].fields[fieldIndex].comment = e.target.value;
                            setConfig({ ...config, tables: newTables });
                          }}
                          className="input-field flex-1 min-w-[150px] h-9 px-3 text-sm"
                          placeholder="注释说明"
                        />
                        {table.fields.length > 1 && (
                          <button onClick={() => removeField(tableIndex, fieldIndex)} className="p-1.5 hover:bg-red-100 dark:hover:bg-red-900/30 rounded transition-colors" title="删除字段">
                            <svg className="w-4 h-4 text-red-400" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                              <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M6 18L18 6M6 6l12 12" />
                            </svg>
                          </button>
                        )}
                      </div>
                    ))}
                    {table.fields.length === 0 && (
                      <div className="text-gray-400 text-sm italic p-3 bg-gray-100 dark:bg-gray-800 rounded-lg">
                        暂无字段，请点击"添加字段"按钮
                      </div>
                    )}
                  </div>
                </div>
              ))}
              {config.tables.length === 0 && (
                <div className="text-center py-12 text-gray-400">
                  <div className="text-5xl mb-4">📊</div>
                  <p>暂无数据表，请点击上方按钮添加</p>
                </div>
              )}
            </div>
          </GlassCard>

          <div className="flex gap-3">
            <button
              onClick={handleGenerate}
              disabled={generating || !config.project_name}
              className="flex-1 btn-primary py-3 disabled:opacity-50"
            >
              {generating ? (
                <span className="flex items-center justify-center gap-2">
                  <svg className="w-5 h-5 animate-spin" fill="none" viewBox="0 0 24 24">
                    <circle className="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" strokeWidth="4" />
                    <path className="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4z" />
                  </svg>
                  {t('generator.generating')}
                </span>
              ) : (
                t('generator.generate')
              )}
            </button>
            {editProjectId && (
              <button
                onClick={handleRegenerate}
                disabled={generating}
                className="flex-1 btn-secondary py-3 disabled:opacity-50"
              >
                {generating ? (
                  <span className="flex items-center justify-center gap-2">
                    <svg className="w-5 h-5 animate-spin" fill="none" viewBox="0 0 24 24">
                      <circle className="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" strokeWidth="4" />
                      <path className="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4z" />
                    </svg>
                    {t('generator.regenerating') || '重新生成中...'}
                  </span>
                ) : (
                  t('projects.regenerate')
                )}
              </button>
            )}
          </div>
        </div>

        <div className="space-y-6">
          <GlassCard>
            <h3 className="text-lg font-semibold text-gray-900 dark:text-white mb-4">{t('generator.generatedStructure')}</h3>
            <div className="space-y-3 text-sm">
              {['go.mod', 'main.go', 'config/', 'database/', 'internal/models/', 'internal/dao/', 'internal/service/', 'internal/controller/', 'internal/router/', 'migrations/'].map((item) => (
                <div key={item} className="flex items-center gap-2 text-gray-600 dark:text-gray-400">
                  <svg className="w-4 h-4 text-primary-500" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                    <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M9 12h6m-6 4h6m2 5H7a2 2 0 01-2-2V5a2 2 0 012-2h5.586a1 1 0 01.707.293l5.414 5.414a1 1 0 01.293.707V19a2 2 0 01-2 2z" />
                  </svg>
                  {item}
                </div>
              ))}
            </div>
          </GlassCard>

          <StatCard title={t('generator.estimatedFiles')} value="25+" />
          <StatCard title={t('generator.linesOfCode')} value="~2,500" />
        </div>

        {result && (
          <GlassCard>
            <h3 className="text-lg font-semibold text-gray-900 dark:text-white mb-4">{t('generator.generationResult')}</h3>
            <div className="space-y-4">
              <div className="flex justify-between items-center flex-wrap gap-3">
                <span className="text-sm text-gray-600 dark:text-gray-400">
                  生成了 <strong>{result?.code?.files ? Object.keys(result.code.files).length : 0}</strong> 个文件
                </span>
                <div className="flex gap-2">
                  {result?.project_id && (
                    <button
                      className="btn-primary text-sm flex items-center gap-2"
                      onClick={handleDownload}
                    >
                      <svg className="w-4 h-4" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                        <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M4 16v1a3 3 0 003 3h10a3 3 0 003-3v-1m-4-4l-4 4m0 0l-4-4m4 4V4" />
                      </svg>
                      {t('common.download')}
                    </button>
                  )}
                  <button
                    className="btn-secondary text-sm flex items-center gap-2"
                    onClick={() => navigate('/projects')}
                  >
                    <svg className="w-4 h-4" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                      <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M13 16h-1v-4h-1l1-2h2l1 2h-1v4zM12 2L2 7l10 5 10-5-10-5z" />
                    </svg>
                    {t('projects.viewInProjects') || '在项目管理中查看'}
                  </button>
                </div>
              </div>
              <div className="grid grid-cols-1 md:grid-cols-2 gap-4 max-h-[300px] overflow-y-auto">
                {result?.code?.files ? Object.keys(result.code.files).map((file, index) => (
                  <div key={index} className="p-3 bg-gray-50 dark:bg-gray-800/50 rounded-lg flex items-center gap-2">
                    <svg className="w-4 h-4 text-primary-500 shrink-0" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                      <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M9 12h6m-6 4h6m2 5H7a2 2 0 01-2-2V5a2 2 0 012-2h5.586a1 1 0 01.707.293l5.414 5.414a1 1 0 01.293.707V19a2 2 0 01-2 2z" />
                    </svg>
                    <span className="text-sm font-medium text-gray-900 dark:text-white truncate">
                      {file}
                    </span>
                  </div>
                )) : <p className="text-gray-500">没有生成的文件</p>}
              </div>
            </div>
          </GlassCard>
        )}
      </div>
    </div>
  );
}
