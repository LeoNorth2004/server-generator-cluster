import { useState, useEffect, useCallback } from 'react';
import { useTranslation } from 'react-i18next';
import { GlassCard, StatCard } from '../components/Cards';

export default function Docs() {
  const [activeTab, setActiveTab] = useState('api');
  const [projects, setProjects] = useState([]);
  const [selectedProjectId, setSelectedProjectId] = useState('');
  const [generatedDoc, setGeneratedDoc] = useState('');
  const [generating, setGenerating] = useState(false);
  const { t } = useTranslation();

  useEffect(() => {
    fetchProjects();
  }, []);

  const fetchProjects = async () => {
    try {
      const response = await fetch('/api/v1/projects', {
        headers: { Authorization: `Bearer ${localStorage.getItem('token')}` }
      });
      if (response.ok) {
        const data = await response.json();
        const projectsData = data?.data || data || [];
        setProjects(Array.isArray(projectsData) ? projectsData : []);
      }
    } catch (error) {
      console.error('Failed to fetch projects:', error);
    }
  };

  const getSelectedProject = useCallback(() => {
    if (!selectedProjectId) return null;
    return projects.find(p => p.id === parseInt(selectedProjectId));
  }, [selectedProjectId, projects]);

  const generateApiDoc = () => {
    const project = getSelectedProject();
    if (!project) {
      alert(t('docs.selectProjectHint'));
      return;
    }

    setActiveTab('api');
    setGenerating(true);
    setGeneratedDoc('');

    setTimeout(() => {
      let doc = `# ${t('docs.apiDocTitle')}\n\n`;
      doc += `> **${t('docs.project')}:** \`${project.name}\`\n`;
      doc += `> **${t('docs.generatedAt')}:** ${new Date().toLocaleString()}\n\n`;
      doc += `---\n\n`;

      let tables = [];
      try {
        const tableConfig = project.table_config ? JSON.parse(project.table_config) : null;
        tables = Array.isArray(tableConfig) ? tableConfig : [];
      } catch (e) {
        console.error('Failed to parse table_config:', e);
        tables = [];
      }

      const baseUrl = window.location.origin;

      doc += `## 📋 ${t('docs.projectOverview')}\n\n`;
      doc += `| Field | Value |\n|--------|-------|\n`;
      doc += `| ${t('docs.projectName')} | \`${project.name}\` |\n`;
      doc += `| ${t('docs.projectDesc')} | ${project.description || '-'} |\n`;
      doc += `| ${t('docs.tablesCount')} | \`${tables.length}\` |\n`;
      doc += `| Base URL | \`${baseUrl}/api/v1\` |\n\n`;

      if (tables.length > 0) {
        doc += `## 🗄️ ${t('docs.tableStructures')}\n\n`;
        tables.forEach((table, idx) => {
          doc += `### ${idx + 1}. \`${table.name}\`\n\n`;
          if (table.comment) {
            doc += `> **${t('docs.tableComment')}:** ${table.comment}\n\n`;
          }
          doc += `| Column | Type | PK | Nullable | Comment |\n|--------|------|----|----------|----------|\n`;
          if (Array.isArray(table.fields)) {
            table.fields.forEach(col => {
              doc += `| \`${col.name}\` | \`${col.type}\` | ${col.primary ? '✅' : '-'} | ${col.nullable ? 'YES' : 'NO'} | ${col.comment || '-'} |\n`;
            });
          }
          doc += `\n`;
        });

        doc += `## 🔗 API Endpoints\n\n`;
        doc += `> ⚠️ All endpoints require **Authorization: Bearer <token>** header except where noted.\n\n`;

        tables.forEach((table, idx) => {
          const entityName = table.name.toLowerCase().replace(/[_-]/g, '');
          const pluralName = entityName + 's';
          const displayName = table.name + 's';

          doc += `---\n\n### ${idx + 1}. ${displayName} APIs\n\n`;

          doc += `#### GET List all ${displayName}\n\n`;
          doc += `\`\`\`http\nGET ${baseUrl}/api/v1/${pluralName}\nAuthorization: Bearer <your-token>\n\`\`\`\n\n`;
          doc += `**Response:** \`200 OK\`\n\n`;
          doc += `\`\`\`json\n{\n  "code": 200,\n  "message": "success",\n  "data": [\n    {\n      "id": 1,\n${Array.isArray(table.fields) && table.fields.length > 0 ? table.fields.slice(0, 3).map(f => `      "${f.name}": "${f.type.includes('int') ? 0 : 'value'}",`).join('\n') : ''}\n      "created_at": "2026-04-07T00:00:00Z"\n    }\n  ]\n}\n\`\`\`\n\n`;

          doc += `#### Get ${table.name} by ID\n\n`;
          doc += `\`\`\`http\nGET ${baseUrl}/api/v1/${pluralName}/:id\nAuthorization: Bearer <your-token>\n\`\`\`\n\n`;
          doc += `**Response:** \`200 OK\`\n\n`;

          doc += `#### Create new ${table.name}\n\n`;
          doc += `\`\`\`http\nPOST ${baseUrl}/api/v1/${pluralName}\nContent-Type: application/json\nAuthorization: Bearer <your-token>\n\n{\n${Array.isArray(table.fields) ? table.fields.filter(f => !f.primary).slice(0, 4).map(f => `  "${f.name}": ${f.type.includes('int') ? '1' : f.type.includes('bool') ? 'true' : f.type.includes('time') ? '"2026-04-07T00:00:00Z"' : '"sample_value"'},`).join('\n') : ''}\n}\n\`\`\`\n\n`;
          doc += `**Response:** \`201 Created\`\n\n`;

          doc += `#### Update ${table.name}\n\n`;
          doc += `\`\`\`http\nPUT ${baseUrl}/api/v1/${pluralName}/:id\nContent-Type: application/json\nAuthorization: Bearer <your-token>\n\n{\n${Array.isArray(table.fields) ? table.fields.filter(f => !f.primary).slice(0, 3).map(f => `  "${f.name}": ${f.type.includes('int') ? '1' : '"new_value"'},`).join('\n') : ''}\n}\n\`\`\`\n\n`;
          doc += `**Response:** \`200 OK\`\n\n`;

          doc += `#### Delete ${table.name}\n\n`;
          doc += `\`\`\`http\nDELETE ${baseUrl}/api/v1/${pluralName}/:id\nAuthorization: Bearer <your-token>\n\`\`\`\n\n`;
          doc += `**Response:** \`200 OK\`\n\n`;
        });
      }

      doc += `\n---\n*${t('docs.autoGenerated')}*`;
      setGeneratedDoc(doc);
      setGenerating(false);
    }, 800);
  };

  const generateConfigDoc = () => {
    const project = getSelectedProject();
    if (!project) {
      alert(t('docs.selectProjectHint'));
      return;
    }

    setActiveTab('config');
    setGenerating(true);
    setGeneratedDoc('');

    setTimeout(() => {
      let doc = `# ${t('docs.configDocTitle')} - ${project.name}\n\n`;
      doc += `> **${t('docs.project')}:** \`${project.name}\`\n`;
      doc += `> **${t('docs.generatedAt')}:** ${new Date().toLocaleString()}\n\n`;
      doc += `---\n\n`;

      let tables = [];
      try {
        const tableConfig = project.table_config ? JSON.parse(project.table_config) : null;
        tables = Array.isArray(tableConfig) ? tableConfig : [];
      } catch (e) {
        console.error('Failed to parse table_config:', e);
        tables = [];
      }

      let dbConfig = {};
      try {
        dbConfig = project.db_config ? JSON.parse(project.db_config) : {};
      } catch (e) {
        console.error('Failed to parse db_config:', e);
      }

      doc += `## 🌐 Environment Variables (.env)\n\n`;
      doc += `\`\`\`bash\n# Application Server\nGIN_MODE=release\nPORT=8080\n\n# Database Configuration\nDB_HOST=${dbConfig.host || 'postgres'}\nDB_PORT=${dbConfig.port || '5432'}\nDB_USER=${dbConfig.user || 'postgres'}\nDB_PASSWORD=${dbConfig.password || '****'}\nDB_NAME=${dbConfig.db_name || 'generator_platform'}\n\n# Redis Configuration\nREDIS_HOST=redis\nREDIS_PORT=6379\nREDIS_PASSWORD=\n\n# JWT Configuration\nJWT_SECRET=<your-secret-key-change-in-production>\nJWT_EXPIRY=24h\n\`\`\`\n\n`;

      doc += `## ⚙️ YAML Configuration (config.yaml)\n\n`;
      doc += `\`\`\`yaml\nserver:\n  host: 0.0.0.0\n  port: 8080\n  mode: release\n\ndatabase:\n  host: ${dbConfig.host || 'postgres'}\n  port: ${parseInt(dbConfig.port) || 5432}\n  user: ${dbConfig.user || 'postgres'}\n  password: ${dbConfig.password || '****'}\n  dbname: ${dbConfig.db_name || 'generator_platform'}\n  sslmode: disable\n  max_connections: 100\n  max_idle: 10\n  conn_max_lifetime: 1h\n\nredis:\n  host: redis\n  port: 6379\n  password: ""\n  db: 0\n  pool_size: 10\n\njwt:\n  secret: <your-secret-key>\n  expiry: 24h\n  issuer: generator-platform\n\ncors:\n  allowed_origins:\n    - "*"\n  allowed_methods:\n    - GET\n    - POST\n    - PUT\n    - DELETE\n    - OPTIONS\n  allowed_headers:\n    - Content-Type\n    - Authorization\n\nlogging:\n  level: info\n  format: json\n\`\`\`\n\n`;

      if (tables.length > 0) {
        doc += `## 📊 Database Schema for Project: ${project.name}\n\n`;
        doc += `Total Tables: **${tables.length}**\n\n`;

        tables.forEach((table, idx) => {
          doc += `### Table ${idx + 1}: \`${table.name}\`\n\n`;
          if (table.comment) {
            doc += `_${table.comment}_\n\n`;
          }
          doc += `| Column | Type | Constraints | Description |\n|--------|------|-------------|-------------|\n`;
          if (Array.isArray(table.fields)) {
            table.fields.forEach(col => {
              const constraints = [];
              if (col.primary) constraints.push('PRIMARY KEY');
              if (!col.nullable) constraints.push('NOT NULL');
              doc += `| \`${col.name}\` | \`${col.type}\` | ${constraints.join(', ') || '-'} | ${col.comment || '-'} |\n`;
            });
          }
          doc += `\n`;

          doc += `**SQL DDL:**\n\n`;
          doc += `\`\`\`sql\nCREATE TABLE IF NOT EXISTS ${table.name} (\n`;
          if (Array.isArray(table.fields)) {
            table.fields.forEach((col, i) => {
              const constraints = [];
              if (col.primary) constraints.push('PRIMARY KEY');
              if (!col.nullable) constraints.push('NOT NULL');
              const comma = i < table.fields.length - 1 ? ',' : '';
              doc += `  ${col.name} ${col.type}${constraints.length > 0 ? ' ' + constraints.join(' ') : ''}${comma}\n`;
            });
          }
          doc += `);\n\`\`\`\n\n`;
        });
      }

      doc += `## ☸️ Kubernetes Deployment Config\n\n`;
      doc += `### Service Discovery\n\n`;
      doc += `In K8S environment, services communicate via DNS:\n\n`;
      doc += `- PostgreSQL: \`${dbConfig.host || 'postgres'}.generator-platform.svc.cluster.local:5432\`\n`;
      doc += `- Redis: \`redis.generator-platform.svc.cluster.local:6379\`\n`;
      doc += `- API Gateway: \`api-gateway.generator-platform.svc.cluster.local:8080\`\n\n`;

      doc += `### Environment Variables for K8S\n\n`;
      doc += `\`\`\`yaml\nenv:\n  - name: DB_HOST\n    value: "postgres.generator-platform.svc.cluster.local"\n  - name: REDIS_HOST\n    value: "redis.generator-platform.svc.cluster.local"\n  - name: PORT\n    value: "8080"\n\`\`\`\n\n`;

      doc += `---\n*${t('docs.autoGenerated')}*`;
      setGeneratedDoc(doc);
      setGenerating(false);
    }, 800);
  };

  const generateDevGuide = () => {
    const project = getSelectedProject();
    if (!project) {
      alert(t('docs.selectProjectHint'));
      return;
    }

    setActiveTab('devguide');
    setGenerating(true);
    setGeneratedDoc('');

    setTimeout(() => {
      let tables = [];
      try {
        const tableConfig = project.table_config ? JSON.parse(project.table_config) : null;
        tables = Array.isArray(tableConfig) ? tableConfig : [];
      } catch (e) {
        console.error('Failed to parse table_config:', e);
        tables = [];
      }

      let doc = `# ${project.name} - ${t('docs.devGuideTitle')}\n\n`;
      doc += `${t('docs.devGuideSubtitle')}\n`;
      doc += `> **${t('docs.project')}:** \`${project.name}\`\n`;
      doc += `> **${t('docs.generatedAt')}:** ${new Date().toLocaleString()}\n`;
      doc += `---\n`;

      doc += `## ${t('docs.projectOverview')}\n\n`;
      doc += `- **${t('docs.projectName')}**: \`${project.name}\`\n`;
      doc += `- **${t('docs.tablesCount')}**: \`${tables.length}\`\n`;
      doc += `- **${t('docs.techStack').split(':')[0]}**: Go + Gin + GORM + PostgreSQL\n\n`;

      if (tables.length > 0) {
        doc += `## 🗄️ ${t('docs.tableStructures')}\n\n`;
        doc += `${t('docs.devGuideContainsTables')}:\n\n`;

        tables.forEach((table, idx) => {
          doc += `### ${idx + 1}. ${table.name}`;
          if (table.comment) { doc += ` (${table.comment})`; }
          doc += `\n\n`;

          if (Array.isArray(table.fields) && table.fields.length > 0) {
            doc += `| ${t('docs.fieldName')} | Type | ${t('docs.nullable')} | PK | ${t('docs.comment')} |\n`;
            doc += `|--------|------|--------|----|----------|\n`;
            table.fields.forEach(field => {
              const primary = field.primary ? '✓' : '';
              const nullable = field.nullable ? t('docs.yes') : t('docs.no');
              const comment = field.comment || '-';
              doc += `| \`${field.name}\` | \`${field.type}\` | ${nullable} | ${primary} | ${comment} |\n`;
            });
            doc += `\n`;
          }
        });

        doc += `## 📁 ${t('docs.projectLayout')}\n\n`;
        doc += `\`\`\`\n${project.name}/\n├── config/                 # ${t('docs.configFile')}\n├── database/               # ${t('docs.dbConnection')}\n├── internal/               # ${t('docs.internalCode')}\n`;

        tables.forEach((table, i) => {
          const modelName = table.name.toLowerCase().replace(/[^a-z0-9]/g, '');
          if (i === 0) {
            doc += `│  ├── models/             # ${t('docs.dataModel')}\n`;
          }
          doc += `│  │  └── ${modelName}.go           # ${table.name} ${t('docs.modelSuffix')}\n`;
        });

        doc += `│  ├── controller/         # ${t('docs.controllerLayer')}\n`;
        doc += `│  ├── dao/                # ${t('docs.daoLayer')}\n`;
        doc += `│  ├── middleware/         # ${t('docs.middleware')}\n`;
        doc += `│  ├── router/             # ${t('docs.routerConfig')}\n`;
        doc += `│  └── service/            # ${t('docs.serviceLayer')}\n`;
        doc += `├── pkg/                    # ${t('docs.commonPkg')}\n`;
        doc += `│  └── utils/              # ${t('docs.utilsFunc')}\n`;
        doc += `├── docs/                   # ${t('docs.docsDir')}\n`;
        doc += `├── migrations/             # ${t('docs.migrationsDir')}\n`;
        doc += `├── go.mod                  # ${t('docs.goModDef')}\n`;
        doc += `├── .env.example            # ${t('docs.envExample')}\n`;
        doc += `└── README.md               # ${t('docs.readmeFile')}\n`;
        doc += `\`\`\`\n\n`;

        doc += `## 💡 ${t('docs.modelExample')}\n\n`;

        tables.slice(0, 1).forEach(table => {
          const modelName = table.name.charAt(0).toUpperCase() + table.name.slice(1).replace(/[^a-zA-Z0-9]/g, '');
          const fileName = table.name.toLowerCase().replace(/[^a-z0-9]/g, '');

          doc += `### internal/models/${fileName}.go\n\n`;
          doc += `\`\`\`go\npackage models\n\nimport "time"\n\ntype ${modelName} struct {\n`;

          if (Array.isArray(table.fields)) {
            table.fields.forEach(field => {
              let goType = 'string';
              if (field.type.toLowerCase().includes('int')) { goType = 'uint'; }
              else if (field.type.toLowerCase().includes('bool')) { goType = 'bool'; }
              else if (field.type.toLowerCase().includes('time') || field.type.toLowerCase().includes('timestamp')) { goType = 'time.Time'; }
              else if (field.type.toLowerCase().includes('text') || field.type.toLowerCase().includes('varchar')) { goType = 'string'; }
              else if (field.type.toLowerCase().includes('decimal') || field.type.toLowerCase().includes('numeric')) { goType = 'float64'; }

              const gormTag = field.primary ? 'primaryKey' : (field.nullable ? '' : 'not null');
              const jsonTag = field.name;

              doc += `    ${fieldNameToCamel(field.Name || field.name)} ${goType} \`gorm:"${gormTag}" json:"${jsonTag}"\`\n`;
            });
          }

          doc += `    CreatedAt time.Time \`json:"created_at"\`\n`;
          doc += `    UpdatedAt time.Time \`json:"updated_at"\`\n`;
          doc += `}\n\n`;
          doc += `func (${modelName}) TableName() string {\n`;
          doc += `    return "${table.name}"\n`;
          doc += `}\n`;
          doc += `\`\`\`\n\n`;
        });

        doc += `## 🔌 ${t('docs.handlerPattern')}\n\n`;

        tables.slice(0, 1).forEach(table => {
          const modelName = table.name.charAt(0).toUpperCase() + table.name.slice(1).replace(/[^a-zA-Z0-9]/g, '');
          const entityLower = table.name.toLowerCase().replace(/[^a-z0-9]/g, '');
          const pluralName = entityLower + 's';

          doc += `### internal/controller/${entityLower}.go\n\n`;
          doc += `\`\`\`go\npackage controller\n\nimport (\n    "net/http"\n    "your-project/internal/models"\n    "your-project/pkg/response"\n\n    "github.com/gin-gonic/gin"\n)\n\ntype ${modelName}Controller struct {}\n\nfunc New${modelName}Controller() *${modelName}Controller {\n    return &${modelName}Controller{}\n}\n\n// Create - POST /api/v1/${pluralName}\nfunc (ctrl *${modelName}Controller) Create(c *gin.Context) {\n    var req models.${modelName}\n    if err := c.ShouldBindJSON(&req); err != nil {\n        response.BadRequest(c, err.Error())\n        return\n    }\n\n    // TODO: 调用 service 层创建记录\n    response.Created(c, req)\n}\n\n// GetByID - GET /api/v1/${pluralName}/:id\nfunc (ctrl *${modelName}Controller) GetByID(c *gin.Context) {\n    id := c.Param("id")\n    // TODO: 调用 service 层获取记录\n    response.Success(c, gin.H{"id": id})\n}\n\n// List - GET /api/v1/${pluralName}\nfunc (ctrl *${modelName}Controller) List(c *gin.Context) {\n    // TODO: 调用 service 层获取列表\n    response.Success(c, []models.${modelName}{})\n}\n\n// Update - PUT /api/v1/${pluralName}/:id\nfunc (ctrl *${modelName}Controller) Update(c *gin.Context) {\n    id := c.Param("id")\n    var req models.${modelName}\n    if err := c.ShouldBindJSON(&req); err != nil {\n        response.BadRequest(c, err.Error())\n        return\n    }\n    // TODO: 调用 service 层更新记录\n    response.Success(c, gin.H{"id": id})\n}\n\n// Delete - DELETE /api/v1/${pluralName}/:id\nfunc (ctrl *${modelName}Controller) Delete(c *gin.Context) {\n    id := c.Param("id")\n    // TODO: 调用 service 层删除记录\n    response.Success(c, gin.H{"message": "deleted"})\n}\n`;
          doc += `\`\`\`\n\n`;
        });

        doc += `## 🔗 API Endpoints\n\n`;
        doc += `${t('docs.generatedApiEndpoints')}:\n\n`;

        tables.forEach(table => {
          const entityLower = table.name.toLowerCase().replace(/[^a-z0-9]/g, '');
          const pluralName = entityLower + 's';
          const modelName = table.name.charAt(0).toUpperCase() + table.name.slice(1).replace(/[^a-zA-Z0-9]/g, '');

          doc += `### ${table.name} APIs\n\n`;
          doc += `| Method | Endpoint | Handler |\n`;
          doc += `|--------|----------|----------|\n`;
          doc += `| POST | \`/api/v1/${pluralName}\` | Create${modelName} |\n`;
          doc += `| GET | \`/api/v1/${pluralName}/:id\` | Get${modelName}ByID |\n`;
          doc += `| GET | \`/api/v1/${pluralName}\` | List${modelName}s |\n`;
          doc += `| PUT | \`/api/v1/${pluralName}/:id\` | Update${modelName} |\n`;
          doc += `| DELETE | \`/api/v1/${pluralName}/:id\` | Delete${modelName} |\n\n`;
        });
      }

      doc += `## 🚀 Quick Start Guide\n\n`;
      doc += `### 1. ${t('docs.envPreparation')}\n\n`;
      doc += `\`\`\`bash\n# ${t('docs.installDeps')}\ngo mod download\n\n# ${t('docs.copyConfig')}\ncp .env.example .env\n\`\`\`\n\n`;

      doc += `### 2. ${t('docs.dbMigration')}\n\n`;
      doc += `\`\`\`bash\n# ${t('docs.runMigrationCmd')}\n./server migrate\n\`\`\`\n\n`;

      doc += `### 3. ${t('docs.startService')}\n\n`;
      doc += `\`\`\`bash\n# ${t('docs.devMode')}\nGIN_MODE=debug ./server serve\n\n# ${t('docs.prodMode')}\n./server serve\n\`\`\`\n\n`;

      doc += `### 4. ${t('docs.dockerDeploy')}\n\n`;
      doc += `\`\`\`bash\ndocker build -t ${project.name.toLowerCase()} .\ndocker run -d -p 8080:8080 --env-file .env ${project.name.toLowerCase()}\n\`\`\`\n\n`;

      doc += `---\n*${t('docs.autoGenerated')}*`;
      setGeneratedDoc(doc);
      setGenerating(false);
    }, 800);
  };

  const fieldNameToCamel = (name) => {
    if (!name) return name;
    const parts = name.split('_');
    return parts.map(part => part.charAt(0).toUpperCase() + part.slice(1).toLowerCase()).join('');
  };

  const handleCopyDoc = () => {
    navigator.clipboard.writeText(generatedDoc).then(() => {
      alert(t('docs.copied'));
    });
  };

  const handleDownloadDoc = () => {
    const blob = new Blob([generatedDoc], { type: 'text/markdown' });
    const url = URL.createObjectURL(blob);
    const a = document.createElement('a');
    a.href = url;
    a.download = activeTab === 'api' ? 'API-Doc.md' : activeTab === 'config' ? 'Config-Doc.md' : 'DevGuide.md';
    document.body.appendChild(a);
    a.click();
    document.body.removeChild(a);
    URL.revokeObjectURL(url);
  };

  const tabs = [
    { id: 'api', label: t('docs.apiDoc'), icon: '📡' },
    { id: 'config', label: t('docs.configDoc'), icon: '⚙️' },
    { id: 'devguide', label: t('docs.devGuide'), icon: '🔧' },
  ];

  return (
    <div className="space-y-6">
      <div className="flex items-center justify-between">
        <div>
          <h1 className="text-2xl font-bold text-gray-900 dark:text-white">{t('docs.title')}</h1>
          <p className="text-gray-500 dark:text-gray-400 mt-1">{t('docs.subtitle')}</p>
        </div>
        <div className="flex gap-2">
          {generatedDoc && (
            <>
              <button onClick={handleCopyDoc} className="btn-secondary flex items-center gap-2">
                📋 {t('docs.copy')}
              </button>
              <button onClick={handleDownloadDoc} className="btn-primary flex items-center gap-2">
                ⬇️ {t('common.download')}
              </button>
            </>
          )}
        </div>
      </div>

      <GlassCard>
        <div className="flex gap-2 border-b border-gray-200 dark:border-gray-700 overflow-x-auto mb-6">
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

        {(activeTab === 'api' || activeTab === 'config' || activeTab === 'devguide') && (
          <div className="space-y-4 mb-6 p-4 bg-gray-50 dark:bg-gray-800 rounded-lg">
            <div className="flex items-start gap-4">
              <div className="flex-1">
                <label className="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-2">
                  {t('docs.selectProject')}
                </label>
                <select
                  value={selectedProjectId}
                  onChange={(e) => setSelectedProjectId(e.target.value)}
                  className="input-field w-full"
                >
                  <option value="">{t('docs.selectProjectPlaceholder')}</option>
                  {projects.map(p => (
                    <option key={p.id} value={p.id}>{p.name}</option>
                  ))}
                </select>
              </div>
              <div className="pt-7">
                <button
                  onClick={() => {
                    if (activeTab === 'api') generateApiDoc();
                    else if (activeTab === 'config') generateConfigDoc();
                    else generateDevGuide();
                  }}
                  disabled={!selectedProjectId || generating}
                  className="btn-primary whitespace-nowrap disabled:opacity-50 disabled:cursor-not-allowed"
                >
                  {generating ? `${t('docs.generating')}...` : (activeTab === 'api' ? t('docs.generateApiDoc') : activeTab === 'config' ? t('docs.generateConfigDoc') : t('docs.generateDevGuide'))}
                </button>
              </div>
            </div>
            {!selectedProjectId && (
              <p className="text-sm text-yellow-600 dark:text-yellow-400 mt-2">
                ⚠️ {t('docs.selectProjectHint')}
              </p>
            )}
          </div>
        )}

        {activeTab === 'devguide' && !generatedDoc && !selectedProjectId && (
          <div className="text-center py-12">
            <div className="text-6xl mb-4">🔧</div>
            <h3 className="text-xl font-semibold mb-2">{t('docs.devGuideTitle')}</h3>
            <p className="text-gray-500 max-w-md mx-auto">{t('docs.devGuideDesc')}</p>
          </div>
        )}

        {generating && (
          <div className="flex items-center justify-center py-20">
            <svg className="w-10 h-10 animate-spin text-primary-600" fill="none" viewBox="0 0 24 24">
              <circle className="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" strokeWidth="4" />
              <path className="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4z" />
            </svg>
            <span className="ml-3 text-gray-600">{t('docs.generating')}</span>
          </div>
        )}

        {generatedDoc && !generating && (
          <div className="bg-gray-50 dark:bg-gray-900 rounded-lg p-6 overflow-auto max-h-[70vh]">
            <pre className="text-sm whitespace-pre-wrap text-gray-800 dark:text-gray-200 font-mono leading-relaxed">{generatedDoc}</pre>
          </div>
        )}
      </GlassCard>
    </div>
  );
}
