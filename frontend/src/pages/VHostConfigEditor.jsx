import { useState, useEffect } from 'react';
import { useParams, useNavigate } from 'react-router-dom';
import { ArrowLeft, Save, RotateCcw, Copy, Loader2, CheckCircle, AlertCircle } from 'lucide-react';
import CodeMirror from '@uiw/react-codemirror';
import { monokai } from '../theme/monokai';
import { nginx } from '../lang/nginx';

const VHostConfigEditor = () => {
  const { domain } = useParams();
  const navigate = useNavigate();
  const [loading, setLoading] = useState(true);
  const [saving, setSaving] = useState(false);
  const [error, setError] = useState(null);
  const [success, setSuccess] = useState(null);
  const [config, setConfig] = useState(null);
  const [content, setContent] = useState('');
  const [originalContent, setOriginalContent] = useState('');

  useEffect(() => {
    fetchConfig();
  }, [domain]);

  const fetchConfig = async () => {
    try {
      setLoading(true);
      setError(null);
      const token = localStorage.getItem('token');
      const response = await fetch(`/api/v1/vhost-config/${domain}`, {
        headers: {
          Authorization: `Bearer ${token}`,
        },
      });

      if (!response.ok) {
        throw new Error('Failed to fetch config');
      }

      const data = await response.json();
      setConfig(data);
      setContent(data.content);
      setOriginalContent(data.content);
    } catch (err) {
      setError(err.message || 'Failed to load config');
    } finally {
      setLoading(false);
    }
  };

  const handleSave = async () => {
    try {
      setSaving(true);
      setError(null);
      setSuccess(null);
      const token = localStorage.getItem('token');
      const response = await fetch(`/api/v1/vhost-config/${domain}`, {
        method: 'PUT',
        headers: {
          'Content-Type': 'application/json',
          Authorization: `Bearer ${token}`,
        },
        body: JSON.stringify({ content }),
      });

      if (!response.ok) {
        throw new Error('Failed to save config');
      }

      const result = await response.json();
      setSuccess(result.message || 'Config saved successfully');
      setOriginalContent(content);
      
      // Auto-hide success message after 3 seconds
      setTimeout(() => setSuccess(null), 3000);
    } catch (err) {
      setError(err.message || 'Failed to save config');
    } finally {
      setSaving(false);
    }
  };

  const handleReset = () => {
    setContent(originalContent);
    setSuccess(null);
    setError(null);
  };

  const handleCopy = () => {
    navigator.clipboard.writeText(content);
    setSuccess('Config copied to clipboard');
    setTimeout(() => setSuccess(null), 2000);
  };

  const hasChanges = content !== originalContent;

  if (loading) {
    return (
      <div className="container mx-auto px-4 py-8">
        <div className="flex justify-center items-center h-96">
          <Loader2 className="w-8 h-8 animate-spin text-primary-600" />
        </div>
      </div>
    );
  }

  return (
    <div className="container mx-auto px-4 py-8">
      <div className="mb-6">
        <div className="flex items-center gap-3 mb-4">
          <button
            onClick={() => navigate('/vhosts')}
            className="btn btn-secondary p-2"
          >
            <ArrowLeft className="w-5 h-5" />
          </button>
          <h1 className="text-3xl font-bold text-gray-900">Nginx Config Editor</h1>
        </div>

        {error && (
          <div className="bg-red-50 border border-red-200 text-red-800 px-4 py-3 rounded-lg flex items-start gap-2 mb-4">
            <AlertCircle className="w-5 h-5 flex-shrink-0 mt-0.5" />
            <div className="flex-1">
              <p className="font-medium">Error</p>
              <p className="text-sm">{error}</p>
            </div>
            <button onClick={() => setError(null)} className="text-red-600 hover:text-red-800">
              ×
            </button>
          </div>
        )}

        {success && (
          <div className="bg-green-50 border border-green-200 text-green-800 px-4 py-3 rounded-lg flex items-start gap-2 mb-4">
            <CheckCircle className="w-5 h-5 flex-shrink-0 mt-0.5" />
            <div className="flex-1">
              <p className="font-medium">Success</p>
              <p className="text-sm">{success}</p>
            </div>
            <button onClick={() => setSuccess(null)} className="text-green-600 hover:text-green-800">
              ×
            </button>
          </div>
        )}

        <div className="card">
          <div className="flex justify-between items-center mb-4 pb-4 border-b border-gray-200">
            <div>
              <h2 className="text-xl font-semibold text-gray-900">{domain}</h2>
              <p className="text-sm text-gray-500 mt-1">{config?.path}</p>
            </div>
            <div className="flex items-center gap-2">
              <button
                onClick={handleCopy}
                className="btn btn-secondary flex items-center gap-2"
                title="Copy to clipboard"
              >
                <Copy className="w-4 h-4" />
                <span className="hidden sm:inline">Copy</span>
              </button>
              <button
                onClick={handleReset}
                disabled={!hasChanges}
                className="btn btn-secondary flex items-center gap-2 disabled:opacity-50 disabled:cursor-not-allowed"
                title="Reset changes"
              >
                <RotateCcw className="w-4 h-4" />
                <span className="hidden sm:inline">Reset</span>
              </button>
              <button
                onClick={handleSave}
                disabled={!hasChanges || saving}
                className="btn btn-primary flex items-center gap-2 disabled:opacity-50 disabled:cursor-not-allowed"
              >
                {saving ? (
                  <>
                    <Loader2 className="w-4 h-4 animate-spin" />
                    <span>Saving...</span>
                  </>
                ) : (
                  <>
                    <Save className="w-4 h-4" />
                    <span>Save Config</span>
                  </>
                )}
              </button>
            </div>
          </div>

          <div className="rounded-lg overflow-hidden border border-gray-700">
            <CodeMirror
              value={content}
              height="600px"
              theme={monokai}
              extensions={[nginx]}
              onChange={(value) => setContent(value)}
              basicSetup={{
                lineNumbers: true,
                highlightActiveLineGutter: true,
                highlightSpecialChars: true,
                foldGutter: true,
                drawSelection: true,
                dropCursor: true,
                allowMultipleSelections: true,
                indentOnInput: true,
                bracketMatching: true,
                closeBrackets: true,
                autocompletion: false,
                rectangularSelection: true,
                crosshairCursor: true,
                highlightActiveLine: true,
                highlightSelectionMatches: true,
                closeBracketsKeymap: true,
                searchKeymap: true,
                foldKeymap: true,
                completionKeymap: false,
                lintKeymap: true,
              }}
              style={{
                fontSize: '14px',
                fontFamily: '"Fira Code", "Consolas", "Monaco", monospace',
              }}
            />
          </div>

          <div className="flex justify-between items-center mt-4 pt-4 border-t border-gray-200">
            <p className="text-sm text-gray-600">
              {hasChanges ? (
                <span className="text-yellow-600 font-medium">● Unsaved changes</span>
              ) : (
                <span className="text-green-600">✓ No changes</span>
              )}
            </p>
            <p className="text-xs text-gray-500">
              Note: Config will be backed up automatically before saving
            </p>
          </div>
        </div>
      </div>
    </div>
  );
};

export default VHostConfigEditor;
