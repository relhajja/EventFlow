import { useState } from 'react';
import { GitBranch, Code, Container } from 'lucide-react';

interface GitFormProps {
  onSubmit: (data: { url: string; branch: string; path: string }) => void;
}

interface ImageFormProps {
  onSubmit: (image: string) => void;
}

export function GitForm({ onSubmit }: GitFormProps) {
  const [url, setUrl] = useState('');
  const [branch, setBranch] = useState('main');
  const [path, setPath] = useState('./');

  const handleSubmit = (e: React.FormEvent) => {
    e.preventDefault();
    onSubmit({ url, branch, path });
  };

  return (
    <div className="bg-slate-800 border border-slate-700 rounded-lg p-6">
      <form onSubmit={handleSubmit} className="space-y-4">
        <div>
          <label className="block text-sm font-medium text-slate-300 mb-2">
            Git Repository URL *
          </label>
          <input
            type="url"
            value={url}
            onChange={(e) => setUrl(e.target.value)}
            placeholder="https://github.com/user/repo.git"
            required
            className="w-full px-4 py-2 bg-slate-900 border border-slate-700 rounded-lg text-white placeholder-slate-500 focus:ring-2 focus:ring-primary-500 focus:border-transparent"
          />
        </div>

        <div className="grid grid-cols-2 gap-4">
          <div>
            <label className="block text-sm font-medium text-slate-300 mb-2">
              Branch
            </label>
            <input
              type="text"
              value={branch}
              onChange={(e) => setBranch(e.target.value)}
              placeholder="main"
              className="w-full px-4 py-2 bg-slate-900 border border-slate-700 rounded-lg text-white placeholder-slate-500 focus:ring-2 focus:ring-primary-500 focus:border-transparent"
            />
          </div>

          <div>
            <label className="block text-sm font-medium text-slate-300 mb-2">
              Path
            </label>
            <input
              type="text"
              value={path}
              onChange={(e) => setPath(e.target.value)}
              placeholder="./"
              className="w-full px-4 py-2 bg-slate-900 border border-slate-700 rounded-lg text-white placeholder-slate-500 focus:ring-2 focus:ring-primary-500 focus:border-transparent"
            />
          </div>
        </div>

        <div className="bg-amber-900/20 border border-amber-500/30 rounded-lg p-3">
          <p className="text-sm text-amber-300">
            <strong>Note:</strong> Private repositories are not yet supported. Coming soon!
          </p>
        </div>

        <div className="bg-blue-900/20 border border-blue-500/30 rounded-lg p-3">
          <p className="text-sm text-blue-300">
            <strong>Auto-detection:</strong> Runtime will be automatically detected from your repository:
          </p>
          <ul className="text-sm text-blue-300 mt-2 ml-4 list-disc">
            <li><code className="bg-blue-900/50 px-1 rounded">requirements.txt</code> → Python</li>
            <li><code className="bg-blue-900/50 px-1 rounded">package.json</code> → Node.js</li>
            <li><code className="bg-blue-900/50 px-1 rounded">go.mod</code> → Go</li>
          </ul>
        </div>

        <button
          type="submit"
          disabled={!url.trim()}
          className="w-full px-6 py-3 bg-primary-600 text-white rounded-lg hover:bg-primary-700 disabled:opacity-50 disabled:cursor-not-allowed transition-colors font-medium"
        >
          Deploy from Git Repository
        </button>
      </form>
    </div>
  );
}

export function ImageForm({ onSubmit }: ImageFormProps) {
  const [image, setImage] = useState('');

  const handleSubmit = (e: React.FormEvent) => {
    e.preventDefault();
    onSubmit(image);
  };

  return (
    <div className="bg-slate-800 border border-slate-700 rounded-lg p-6">
      <form onSubmit={handleSubmit} className="space-y-4">
        <div>
          <label className="block text-sm font-medium text-slate-300 mb-2">
            Docker Image *
          </label>
          <input
            type="text"
            value={image}
            onChange={(e) => setImage(e.target.value)}
            placeholder="nginx:alpine"
            required
            className="w-full px-4 py-2 bg-slate-900 border border-slate-700 rounded-lg text-white placeholder-slate-500 focus:ring-2 focus:ring-primary-500 focus:border-transparent"
          />
          <p className="text-sm text-slate-400 mt-2">
            Example: <code className="bg-slate-700 px-1 rounded">nginx:alpine</code>, <code className="bg-slate-700 px-1 rounded">my-registry.com/my-app:v1.0</code>
          </p>
        </div>

        <div className="bg-blue-900/20 border border-blue-500/30 rounded-lg p-3">
          <p className="text-sm text-blue-300">
            <strong>Requirements:</strong> Your image must listen on port 8080 and respond to HTTP requests.
          </p>
        </div>

        <button
          type="submit"
          disabled={!image.trim()}
          className="w-full px-6 py-3 bg-primary-600 text-white rounded-lg hover:bg-primary-700 disabled:opacity-50 disabled:cursor-not-allowed transition-colors font-medium"
        >
          Deploy Docker Image
        </button>
      </form>
    </div>
  );
}

type DeploymentType = 'git' | 'code' | 'image';

interface DeploymentTypeSelectorProps {
  selected: DeploymentType;
  onSelect: (type: DeploymentType) => void;
}

export function DeploymentTypeSelector({ selected, onSelect }: DeploymentTypeSelectorProps) {
  const types = [
    {
      id: 'git' as const,
      icon: GitBranch,
      title: 'Git Repository',
      description: 'Deploy from GitHub, GitLab, or Bitbucket',
      badge: 'Production'
    },
    {
      id: 'code' as const,
      icon: Code,
      title: 'Write Code',
      description: 'Use browser editor with syntax highlighting',
      badge: 'Quick Start'
    },
    {
      id: 'image' as const,
      icon: Container,
      title: 'Docker Image',
      description: 'Use existing container image',
      badge: 'Advanced'
    }
  ];

  return (
    <div className="bg-slate-800 border border-slate-700 rounded-lg p-6">
      <h3 className="text-lg font-semibold text-white mb-4">Select Deployment Method</h3>
      <div className="grid grid-cols-1 md:grid-cols-3 gap-4">
        {types.map((type) => {
          const Icon = type.icon;
          const isSelected = selected === type.id;
          
          return (
            <button
              key={type.id}
              onClick={() => onSelect(type.id)}
              type="button"
              className={`p-4 rounded-lg border-2 transition-all text-left ${
                isSelected
                  ? 'border-primary-500 bg-primary-500/10'
                  : 'border-slate-600 bg-slate-900 hover:border-slate-500'
              }`}
            >
              <div className="flex items-start justify-between mb-2">
                <Icon className={`w-6 h-6 ${isSelected ? 'text-primary-400' : 'text-slate-400'}`} />
                <span className={`text-xs px-2 py-1 rounded-full ${
                  isSelected
                    ? 'bg-primary-500/20 text-primary-300'
                    : 'bg-slate-700 text-slate-400'
                }`}>
                  {type.badge}
                </span>
              </div>
              <h3 className={`font-medium mb-1 ${isSelected ? 'text-white' : 'text-slate-300'}`}>
                {type.title}
              </h3>
              <p className={`text-sm ${isSelected ? 'text-slate-300' : 'text-slate-500'}`}>
                {type.description}
              </p>
            </button>
          );
        })}
      </div>
    </div>
  );
}
