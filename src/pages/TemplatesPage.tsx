import { useState } from 'react';
import { FileEdit, Code, Globe, Server, ArrowRight, X, CheckCircle } from 'lucide-react';

const templates = [
  {
    id: 't1',
    name: 'Go Microservice',
    description: 'Create a new Go microservice with gRPC, health checks, and CI/CD pipeline',
    icon: Server,
    owner: 'Platform Team',
    tags: ['go', 'grpc', 'microservice'],
    language: 'Go',
    repository: 'github.com/company/',
  },
  {
    id: 't2',
    name: 'React Frontend',
    description: 'Scaffold a new React frontend application with TypeScript and testing setup',
    icon: Globe,
    owner: 'Frontend Team',
    tags: ['react', 'typescript', 'frontend'],
    language: 'TypeScript',
    repository: 'github.com/company/',
  },
  {
    id: 't3',
    name: 'Python FastAPI Service',
    description: 'Create a Python service with FastAPI, async support, and OpenAPI docs',
    icon: Code,
    owner: 'Data Team',
    tags: ['python', 'fastapi', 'async'],
    language: 'Python',
    repository: 'github.com/company/',
  },
  {
    id: 't4',
    name: 'Java Spring Boot Service',
    description: 'Spring Boot service with JPA, Flyway migrations, and monitoring',
    icon: Server,
    owner: 'Backend Team',
    tags: ['java', 'spring-boot', 'jpa'],
    language: 'Java',
    repository: 'github.com/company/',
  },
  {
    id: 't5',
    name: 'Kotlin Service',
    description: 'Kotlin service with Spring Boot, coroutines, and gRPC support',
    icon: Server,
    owner: 'Orders Team',
    tags: ['kotlin', 'spring-boot', 'coroutines'],
    language: 'Kotlin',
    repository: 'github.com/company/',
  },
  {
    id: 't6',
    name: 'Release Template',
    description: 'Standard release process template with build, test, and deploy stages',
    icon: FileEdit,
    owner: 'Release Team',
    tags: ['release', 'ci/cd', 'pipeline'],
    language: 'YAML',
    repository: 'github.com/company/',
  },
];

interface Template {
  id: string;
  name: string;
  description: string;
  icon: any;
  owner: string;
  tags: string[];
  language: string;
  repository: string;
}

export default function TemplatesPage() {
  const [selectedTemplate, setSelectedTemplate] = useState<Template | null>(null);
  const [showForm, setShowForm] = useState(false);
  const [formData, setFormData] = useState({
    name: '',
    description: '',
    owner: '',
    repository: '',
  });
  const [creating, setCreating] = useState(false);
  const [created, setCreated] = useState(false);

  const handleChooseTemplate = (template: Template) => {
    setSelectedTemplate(template);
    setShowForm(true);
    setFormData({
      name: '',
      description: template.description,
      owner: template.owner,
      repository: template.repository,
    });
  };

  const handleClose = () => {
    setShowForm(false);
    setSelectedTemplate(null);
    setCreated(false);
  };

  const handleCreate = async () => {
    if (!formData.name) return;
    
    setCreating(true);
    await new Promise(resolve => setTimeout(resolve, 1500));
    setCreating(false);
    setCreated(true);
  };

  return (
    <div>
      <div className="bg-black text-white">
        <div className="max-w-7xl mx-auto px-8 py-8">
          <p className="text-xs text-gray-400 uppercase tracking-wider mb-1">Home / Templates</p>
          <h1 className="text-2xl font-light">Software Templates</h1>
          <p className="text-gray-400 mt-1 text-sm">Create new components using pre-defined templates</p>
        </div>
      </div>

      <div className="max-w-7xl mx-auto px-8 -mt-4 pb-8">
        <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-6">
          {templates.map(template => (
            <div 
              key={template.id} 
              onClick={() => handleChooseTemplate(template)}
              className="bg-white rounded-lg shadow-sm border border-gray-200 overflow-hidden hover:shadow-md transition-shadow cursor-pointer group"
            >
              <div className="p-6">
                <div className="flex items-center gap-3 mb-3">
                  <div className="p-2 bg-orange-50 rounded-lg">
                    <template.icon className="w-5 h-5 text-orange-600" />
                  </div>
                  <h3 className="font-medium text-gray-900">{template.name}</h3>
                </div>
                <p className="text-sm text-gray-600 mb-4">{template.description}</p>
                <div className="flex items-center gap-1 flex-wrap mb-4">
                  {template.tags.map(tag => (
                    <span key={tag} className="text-xs px-2 py-0.5 bg-gray-100 text-gray-600 rounded">{tag}</span>
                  ))}
                </div>
                <div className="flex items-center justify-between">
                  <span className="text-xs text-gray-400">Owner: {template.owner}</span>
                  <span className="text-sm text-orange-500 group-hover:text-orange-600 flex items-center gap-1">
                    Choose <ArrowRight className="w-3 h-3" />
                  </span>
                </div>
              </div>
            </div>
          ))}
        </div>
      </div>

      {/* Modal */}
      {showForm && selectedTemplate && (
        <div className="fixed inset-0 z-50 flex items-center justify-center p-4">
          <div className="absolute inset-0 bg-black/50" onClick={handleClose} />
          <div className="relative bg-white rounded-lg shadow-2xl w-full max-w-2xl max-h-[90vh] overflow-auto">
            {!created ? (
              <>
                <div className="sticky top-0 bg-white border-b border-gray-200 px-6 py-4 flex items-center justify-between">
                  <div className="flex items-center gap-3">
                    <div className="p-2 bg-orange-50 rounded-lg">
                      <selectedTemplate.icon className="w-5 h-5 text-orange-600" />
                    </div>
                    <div>
                      <h2 className="text-lg font-semibold text-gray-900">Create from template</h2>
                      <p className="text-xs text-gray-500">{selectedTemplate.name}</p>
                    </div>
                  </div>
                  <button onClick={handleClose} className="text-gray-400 hover:text-gray-600">
                    <X className="w-5 h-5" />
                  </button>
                </div>

                <div className="p-6 space-y-4">
                  <div>
                    <label className="block text-sm font-medium text-gray-700 mb-1">
                      Component Name *
                    </label>
                    <input
                      type="text"
                      value={formData.name}
                      onChange={(e) => setFormData({ ...formData, name: e.target.value })}
                      placeholder="my-new-service"
                      className="w-full px-3 py-2 border border-gray-300 rounded-lg focus:outline-none focus:ring-2 focus:ring-orange-500"
                    />
                  </div>

                  <div>
                    <label className="block text-sm font-medium text-gray-700 mb-1">
                      Description
                    </label>
                    <textarea
                      value={formData.description}
                      onChange={(e) => setFormData({ ...formData, description: e.target.value })}
                      rows={3}
                      className="w-full px-3 py-2 border border-gray-300 rounded-lg focus:outline-none focus:ring-2 focus:ring-orange-500 resize-none"
                    />
                  </div>

                  <div>
                    <label className="block text-sm font-medium text-gray-700 mb-1">
                      Owner Team
                    </label>
                    <input
                      type="text"
                      value={formData.owner}
                      onChange={(e) => setFormData({ ...formData, owner: e.target.value })}
                      className="w-full px-3 py-2 border border-gray-300 rounded-lg focus:outline-none focus:ring-2 focus:ring-orange-500"
                    />
                  </div>

                  <div>
                    <label className="block text-sm font-medium text-gray-700 mb-1">
                      Repository
                    </label>
                    <input
                      type="text"
                      value={formData.repository}
                      onChange={(e) => setFormData({ ...formData, repository: e.target.value })}
                      placeholder="github.com/company/my-new-service"
                      className="w-full px-3 py-2 border border-gray-300 rounded-lg focus:outline-none focus:ring-2 focus:ring-orange-500"
                    />
                  </div>

                  <div className="bg-gray-50 rounded-lg p-4">
                    <h4 className="text-sm font-medium text-gray-900 mb-2">Template includes:</h4>
                    <ul className="space-y-1 text-xs text-gray-600">
                      <li>✓ Source code structure</li>
                      <li>✓ CI/CD pipeline configuration</li>
                      <li>✓ Docker setup</li>
                      <li>✓ Monitoring and logging</li>
                      <li>✓ Documentation template</li>
                    </ul>
                  </div>
                </div>

                <div className="sticky bottom-0 bg-white border-t border-gray-200 px-6 py-4 flex items-center justify-end gap-3">
                  <button
                    onClick={handleClose}
                    className="px-4 py-2 border border-gray-300 text-gray-700 text-sm font-medium rounded hover:bg-gray-50"
                  >
                    Cancel
                  </button>
                  <button
                    onClick={handleCreate}
                    disabled={!formData.name || creating}
                    className="px-6 py-2 bg-orange-500 hover:bg-orange-600 disabled:bg-gray-300 text-white text-sm font-medium rounded"
                  >
                    {creating ? 'Creating...' : 'Create Component'}
                  </button>
                </div>
              </>
            ) : (
              <div className="p-12 text-center">
                <div className="w-16 h-16 bg-green-100 rounded-full flex items-center justify-center mx-auto mb-4">
                  <CheckCircle className="w-8 h-8 text-green-600" />
                </div>
                <h3 className="text-xl font-semibold text-gray-900 mb-2">Component Created!</h3>
                <p className="text-gray-600 mb-6">
                  Your component <strong>{formData.name}</strong> has been successfully created from the {selectedTemplate.name} template.
                </p>
                <div className="flex items-center justify-center gap-3">
                  <button
                    onClick={handleClose}
                    className="px-4 py-2 border border-gray-300 text-gray-700 text-sm font-medium rounded hover:bg-gray-50"
                  >
                    Close
                  </button>
                  <button
                    onClick={() => window.location.hash = '/catalog'}
                    className="px-6 py-2 bg-orange-500 hover:bg-orange-600 text-white text-sm font-medium rounded"
                  >
                    View in Catalog
                  </button>
                </div>
              </div>
            )}
          </div>
        </div>
      )}
    </div>
  );
}
