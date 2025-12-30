import React, { useState, useEffect } from 'react';
import { Shield, Upload, Plus, Trash2, Eye, EyeOff, AlertCircle, CheckCircle, Calendar, Search, ChevronLeft, ChevronRight } from 'lucide-react';
import api from '../services/api';
import ConfirmModal from '../components/ConfirmModal';

const Certificates = () => {
  const [certificates, setCertificates] = useState([]);
  const [loading, setLoading] = useState(true);
  const [showModal, setShowModal] = useState(false);
  const [showViewModal, setShowViewModal] = useState(false);
  const [selectedCert, setSelectedCert] = useState(null);
  const [error, setError] = useState('');
  const [success, setSuccess] = useState('');
  const [confirmModal, setConfirmModal] = useState({ isOpen: false, onConfirm: null, title: '', message: '' });
  
  // Datatable states
  const [searchTerm, setSearchTerm] = useState('');
  const [currentPage, setCurrentPage] = useState(1);
  const [itemsPerPage, setItemsPerPage] = useState(10);
  
  const [formData, setFormData] = useState({
    name: '',
    cert_content: '',
    key_content: ''
  });

  const [showCertContent, setShowCertContent] = useState(false);
  const [showKeyContent, setShowKeyContent] = useState(false);

  useEffect(() => {
    fetchCertificates();
  }, []);

  const fetchCertificates = async () => {
    try {
      const response = await api.get('/certificates');
      setCertificates(response.data.certificates || []);
    } catch (err) {
      setError('Failed to load certificates');
    } finally {
      setLoading(false);
    }
  };

  const handleFileUpload = async (e, type) => {
    const file = e.target.files[0];
    if (!file) return;

    const reader = new FileReader();
    reader.onload = (event) => {
      const content = event.target.result;
      setFormData(prev => ({
        ...prev,
        [type]: content
      }));
    };
    reader.readAsText(file);
  };

  const handleSubmit = async (e) => {
    e.preventDefault();
    setError('');
    setSuccess('');

    try {
      if (!formData.cert_content || !formData.key_content) {
        setError('Please upload both certificate and key files');
        return;
      }

      await api.post('/certificates', formData);
      setSuccess('Certificate uploaded successfully');
      setShowModal(false);
      setFormData({ name: '', cert_content: '', key_content: '' });
      fetchCertificates();
    } catch (err) {
      setError(err.response?.data?.error || 'Failed to upload certificate');
    }
  };

  const handleDelete = (id) => {
    setConfirmModal({
      isOpen: true,
      title: 'Delete Certificate',
      message: 'Are you sure you want to delete this certificate? This action cannot be undone.',
      type: 'danger',
      onConfirm: async () => {
        try {
          await api.delete(`/certificates/${id}`);
          setSuccess('Certificate deleted successfully');
          fetchCertificates();
        } catch (err) {
          setError('Failed to delete certificate');
        }
      }
    });
  };

  const handleView = async (cert) => {
    try {
      const response = await api.get(`/certificates/${cert.id}`);
      setSelectedCert(response.data);
      setShowViewModal(true);
    } catch (err) {
      setError('Failed to load certificate details');
    }
  };

  const getStatusBadge = (status, validTo) => {
    const now = new Date();
    const expiry = new Date(validTo);
    const daysUntilExpiry = Math.ceil((expiry - now) / (1000 * 60 * 60 * 24));

    let bgColor = 'bg-gray-600';
    let text = status;

    if (status === 'active') {
      bgColor = 'bg-green-600';
      text = 'Active';
    } else if (status === 'expired') {
      bgColor = 'bg-red-600';
      text = 'Expired';
    } else if (status === 'expiring_soon') {
      bgColor = 'bg-yellow-600';
      text = `Expiring in ${daysUntilExpiry} days`;
    } else if (status === 'pending') {
      bgColor = 'bg-blue-600';
      text = 'Pending';
    }

    return (
      <span className={`px-2 py-1 rounded text-xs font-medium text-white ${bgColor}`}>
        {text}
      </span>
    );
  };

  const formatDate = (dateString) => {
    return new Date(dateString).toLocaleDateString('en-US', {
      year: 'numeric',
      month: 'short',
      day: 'numeric',
      hour: '2-digit',
      minute: '2-digit'
    });
  };

  // Filter and pagination
  const filteredCertificates = certificates.filter(cert => {
    const search = searchTerm.toLowerCase();
    return (
      cert.name?.toLowerCase().includes(search) ||
      cert.common_name?.toLowerCase().includes(search) ||
      cert.issuer?.toLowerCase().includes(search) ||
      cert.status?.toLowerCase().includes(search)
    );
  });

  const totalPages = Math.ceil(filteredCertificates.length / itemsPerPage);
  const startIndex = (currentPage - 1) * itemsPerPage;
  const paginatedCertificates = filteredCertificates.slice(startIndex, startIndex + itemsPerPage);

  const handlePageChange = (page) => {
    setCurrentPage(page);
  };

  const handleItemsPerPageChange = (value) => {
    setItemsPerPage(value);
    setCurrentPage(1);
  };

  if (loading) {
    return (
      <div className="flex items-center justify-center h-64">
        <div className="text-gray-400">Loading certificates...</div>
      </div>
    );
  }

  return (
    <div className="space-y-6">
      {/* Header */}
      <div className="flex items-center justify-between">
        <div>
          <h1 className="text-3xl font-bold text-gray-900 mb-2">SSL Certificates</h1>
          <p className="text-gray-600">Manage SSL/TLS certificates for your domains</p>
        </div>
        <button
          onClick={() => setShowModal(true)}
          className="flex items-center gap-2 bg-blue-600 hover:bg-blue-700 text-white px-4 py-2 rounded-lg transition"
        >
          <Plus className="w-5 h-5" />
          Add Certificate
        </button>
      </div>

      {/* Alerts */}
      {error && (
        <div className="p-4 bg-red-50 border border-red-200 rounded-lg flex items-center gap-2">
          <AlertCircle className="w-5 h-5 text-red-600" />
          <p className="text-red-600">{error}</p>
        </div>
      )}

      {success && (
        <div className="p-4 bg-green-50 border border-green-200 rounded-lg flex items-center gap-2">
          <CheckCircle className="w-5 h-5 text-green-600" />
          <p className="text-green-600">{success}</p>
        </div>
      )}

      {/* Search and Filter */}
      <div className="bg-white border rounded-lg p-4 shadow-sm">
        <div className="flex items-center justify-between gap-4">
          <div className="flex-1 relative">
            <Search className="absolute left-3 top-1/2 transform -translate-y-1/2 w-5 h-5 text-gray-400" />
            <input
              type="text"
              placeholder="Search certificates by name, domain, issuer..."
              value={searchTerm}
              onChange={(e) => {
                setSearchTerm(e.target.value);
                setCurrentPage(1);
              }}
              className="w-full pl-10 pr-4 py-2 border border-gray-300 rounded-lg focus:outline-none focus:ring-2 focus:ring-blue-500"
            />
          </div>
          <div className="flex items-center gap-2">
            <label htmlFor="itemsPerPage" className="text-sm text-gray-600">Show:</label>
            <select
              id="itemsPerPage"
              value={itemsPerPage}
              onChange={(e) => handleItemsPerPageChange(Number(e.target.value))}
              className="px-3 py-2 border border-gray-300 rounded-lg focus:outline-none focus:ring-2 focus:ring-blue-500"
            >
              <option value={5}>5</option>
              <option value={10}>10</option>
              <option value={50}>50</option>
              <option value={100}>100</option>
            </select>
          </div>
        </div>
        <div className="mt-2 text-sm text-gray-600">
          Showing {startIndex + 1} to {Math.min(startIndex + itemsPerPage, filteredCertificates.length)} of {filteredCertificates.length} certificates
        </div>
      </div>

      {/* Certificates Table */}
      <div className="bg-white border rounded-lg overflow-hidden shadow-sm">
        <div className="overflow-x-auto">
          <table className="w-full">
            <thead className="bg-gray-50">
              <tr>
                <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
                  Name
                </th>
                <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
                  Common Name
                </th>
                <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
                  Issuer
                </th>
                <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
                  Valid From
                </th>
                <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
                  Valid To
                </th>
                <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
                  Status
                </th>
                <th className="px-6 py-3 text-right text-xs font-medium text-gray-500 uppercase tracking-wider">
                  Actions
                </th>
              </tr>
            </thead>
            <tbody className="bg-white divide-y divide-gray-200">
              {paginatedCertificates.length === 0 ? (
                <tr>
                  <td colSpan="7" className="px-6 py-8 text-center text-gray-500">
                    {searchTerm ? 'No certificates found matching your search.' : 'No certificates found. Upload your first certificate to get started.'}
                  </td>
                </tr>
              ) : (
                paginatedCertificates.map((cert) => (
                  <tr key={cert.id} className="hover:bg-gray-50">
                    <td className="px-6 py-4 whitespace-nowrap">
                      <div className="flex items-center gap-2">
                        <Shield className="w-4 h-4 text-blue-500" />
                        <span className="text-gray-900 font-medium">{cert.name}</span>
                      </div>
                    </td>
                    <td className="px-6 py-4 whitespace-nowrap text-gray-600">
                      {cert.common_name || '-'}
                    </td>
                    <td className="px-6 py-4 whitespace-nowrap text-gray-600">
                      {cert.issuer || '-'}
                    </td>
                    <td className="px-6 py-4 whitespace-nowrap">
                      <div className="flex items-center gap-1 text-gray-500 text-sm">
                        <Calendar className="w-3 h-3" />
                        {formatDate(cert.valid_from)}
                      </div>
                    </td>
                    <td className="px-6 py-4 whitespace-nowrap">
                      <div className="flex items-center gap-1 text-gray-500 text-sm">
                        <Calendar className="w-3 h-3" />
                        {formatDate(cert.valid_to)}
                      </div>
                    </td>
                    <td className="px-6 py-4 whitespace-nowrap">
                      {getStatusBadge(cert.status, cert.valid_to)}
                    </td>
                    <td className="px-6 py-4 whitespace-nowrap text-right">
                      <button
                        onClick={() => handleView(cert)}
                        className="text-blue-600 hover:text-blue-700 mr-3"
                        title="View Details"
                      >
                        <Eye className="w-4 h-4 inline" />
                      </button>
                      <button
                        onClick={() => handleDelete(cert.id)}
                        className="text-red-600 hover:text-red-700"
                        title="Delete"
                      >
                        <Trash2 className="w-4 h-4 inline" />
                      </button>
                    </td>
                  </tr>
                ))
              )}
            </tbody>
          </table>
        </div>

        {/* Pagination */}
        {filteredCertificates.length > 0 && (
          <div className="px-6 py-4 bg-gray-50 border-t border-gray-200">
            <div className="flex items-center justify-between">
              <div className="text-sm text-gray-600">
                Page {currentPage} of {totalPages}
              </div>
              <div className="flex items-center gap-2">
                <button
                  onClick={() => handlePageChange(currentPage - 1)}
                  disabled={currentPage === 1}
                  className="px-3 py-1 border border-gray-300 text-gray-700 rounded hover:bg-gray-50 disabled:opacity-50 disabled:cursor-not-allowed transition"
                >
                  <ChevronLeft className="w-4 h-4" />
                </button>
                
                {Array.from({ length: totalPages }, (_, i) => i + 1).map((page) => {
                  if (
                    page === 1 ||
                    page === totalPages ||
                    (page >= currentPage - 1 && page <= currentPage + 1)
                  ) {
                    return (
                      <button
                        key={page}
                        onClick={() => handlePageChange(page)}
                        className={`px-3 py-1 rounded transition ${
                          page === currentPage
                            ? 'bg-blue-600 text-white'
                            : 'border border-gray-300 text-gray-700 hover:bg-gray-50'
                        }`}
                      >
                        {page}
                      </button>
                    );
                  } else if (page === currentPage - 2 || page === currentPage + 2) {
                    return <span key={page} className="px-2 text-gray-400">...</span>;
                  }
                  return null;
                })}

                <button
                  onClick={() => handlePageChange(currentPage + 1)}
                  disabled={currentPage === totalPages}
                  className="px-3 py-1 border border-gray-300 text-gray-700 rounded hover:bg-gray-50 disabled:opacity-50 disabled:cursor-not-allowed transition"
                >
                  <ChevronRight className="w-4 h-4" />
                </button>
              </div>
            </div>
          </div>
        )}
      </div>

      {/* Add Certificate Modal */}
      {showModal && (
        <div className="fixed inset-0 bg-black/50 flex items-center justify-center z-50 p-4">
          <div className="bg-gray-800 rounded-lg max-w-2xl w-full max-h-[90vh] overflow-y-auto">
            <div className="p-6 border-b border-gray-700">
              <h2 className="text-2xl font-bold text-white">Upload SSL Certificate</h2>
            </div>

            <form onSubmit={handleSubmit} className="p-6 space-y-4">
              <div>
                <label className="block text-sm font-medium text-gray-300 mb-2">
                  Certificate Name *
                </label>
                <input
                  type="text"
                  value={formData.name}
                  onChange={(e) => setFormData({ ...formData, name: e.target.value })}
                  className="w-full px-4 py-2 bg-gray-700 border border-gray-600 rounded-lg text-white focus:outline-none focus:ring-2 focus:ring-blue-500"
                  placeholder="e.g., example.com"
                  required
                />
              </div>

              <div>
                <label className="block text-sm font-medium text-gray-300 mb-2">
                  <Upload className="w-4 h-4 inline mr-2" />
                  Certificate File (.crt, .pem) *
                </label>
                <input
                  type="file"
                  id="cert-file"
                  accept=".crt,.pem,.cer"
                  onChange={(e) => handleFileUpload(e, 'cert_content')}
                  className="w-full px-4 py-2 bg-gray-700 border border-gray-600 rounded-lg text-white file:mr-4 file:py-2 file:px-4 file:rounded file:border-0 file:bg-blue-600 file:text-white hover:file:bg-blue-700"
                  required
                />
              </div>

              {formData.cert_content && (
                <div>
                  <div className="flex items-center justify-between mb-2">
                    <label className="block text-sm font-medium text-gray-300">
                      Certificate Content (Auto-filled)
                    </label>
                    <button
                      type="button"
                      onClick={() => setShowCertContent(!showCertContent)}
                      className="text-gray-400 hover:text-gray-300"
                      title={showCertContent ? "Hide content" : "Show content"}
                    >
                      {showCertContent ? <EyeOff className="w-4 h-4" /> : <Eye className="w-4 h-4" />}
                    </button>
                  </div>
                  <textarea
                    value={formData.cert_content}
                    onChange={(e) => setFormData({ ...formData, cert_content: e.target.value })}
                    className={`w-full px-4 py-2 bg-gray-700 border border-gray-600 rounded-lg text-white font-mono text-sm focus:outline-none focus:ring-2 focus:ring-blue-500 ${!showCertContent ? 'text-security-disc' : ''}`}
                    style={!showCertContent ? { WebkitTextSecurity: 'disc' } : {}}
                    rows="8"
                    placeholder="Certificate content will appear here after file upload"
                  />
                </div>
              )}

              <div>
                <label className="block text-sm font-medium text-gray-300 mb-2">
                  <Upload className="w-4 h-4 inline mr-2" />
                  Private Key File (.key, .pem) *
                </label>
                <input
                  type="file"
                  id="key-file"
                  accept=".key,.pem"
                  onChange={(e) => handleFileUpload(e, 'key_content')}
                  className="w-full px-4 py-2 bg-gray-700 border border-gray-600 rounded-lg text-white file:mr-4 file:py-2 file:px-4 file:rounded file:border-0 file:bg-blue-600 file:text-white hover:file:bg-blue-700"
                  required
                />
              </div>

              {formData.key_content && (
                <div>
                  <div className="flex items-center justify-between mb-2">
                    <label className="block text-sm font-medium text-gray-300">
                      Private Key Content (Auto-filled)
                    </label>
                    <button
                      type="button"
                      onClick={() => setShowKeyContent(!showKeyContent)}
                      className="text-gray-400 hover:text-gray-300"
                      title={showKeyContent ? "Hide content" : "Show content"}
                    >
                      {showKeyContent ? <EyeOff className="w-4 h-4" /> : <Eye className="w-4 h-4" />}
                    </button>
                  </div>
                  <textarea
                    value={formData.key_content}
                    onChange={(e) => setFormData({ ...formData, key_content: e.target.value })}
                    className={`w-full px-4 py-2 bg-gray-700 border border-gray-600 rounded-lg text-white font-mono text-sm focus:outline-none focus:ring-2 focus:ring-blue-500 ${!showKeyContent ? 'text-security-disc' : ''}`}
                    style={!showKeyContent ? { WebkitTextSecurity: 'disc' } : {}}
                    rows="8"
                    placeholder="Private key content will appear here after file upload"
                  />
                </div>
              )}

              <div className="flex gap-3 pt-4">
                <button
                  type="submit"
                  className="flex-1 bg-blue-600 hover:bg-blue-700 text-white py-2 rounded-lg transition"
                >
                  Upload Certificate
                </button>
                <button
                  type="button"
                  onClick={() => {
                    setShowModal(false);
                    setFormData({ name: '', cert_content: '', key_content: '' });
                    setError('');
                  }}
                  className="flex-1 bg-gray-700 hover:bg-gray-600 text-white py-2 rounded-lg transition"
                >
                  Cancel
                </button>
              </div>
            </form>
          </div>
        </div>
      )}

      {/* View Certificate Modal */}
      {showViewModal && selectedCert && (
        <div className="fixed inset-0 bg-black/50 flex items-center justify-center z-50 p-4">
          <div className="bg-gray-800 rounded-lg max-w-3xl w-full max-h-[90vh] overflow-y-auto">
            <div className="p-6 border-b border-gray-700">
              <h2 className="text-2xl font-bold text-white">{selectedCert.name}</h2>
            </div>

            <div className="p-6 space-y-4">
              <div className="grid grid-cols-2 gap-4">
                <div>
                  <label className="block text-sm font-medium text-gray-400 mb-1">Common Name</label>
                  <p className="text-white">{selectedCert.common_name || '-'}</p>
                </div>
                <div>
                  <label className="block text-sm font-medium text-gray-400 mb-1">Issuer</label>
                  <p className="text-white">{selectedCert.issuer || '-'}</p>
                </div>
                <div>
                  <label className="block text-sm font-medium text-gray-400 mb-1">Valid From</label>
                  <p className="text-white">{formatDate(selectedCert.valid_from)}</p>
                </div>
                <div>
                  <label className="block text-sm font-medium text-gray-400 mb-1">Valid To</label>
                  <p className="text-white">{formatDate(selectedCert.valid_to)}</p>
                </div>
                <div>
                  <label className="block text-sm font-medium text-gray-400 mb-1">Status</label>
                  <div>{getStatusBadge(selectedCert.status, selectedCert.valid_to)}</div>
                </div>
                <div>
                  <label className="block text-sm font-medium text-gray-400 mb-1">Created</label>
                  <p className="text-white">{formatDate(selectedCert.created_at)}</p>
                </div>
              </div>

              <div>
                <label className="block text-sm font-medium text-gray-400 mb-2">Certificate Content</label>
                <textarea
                  value={selectedCert.cert_content}
                  readOnly
                  className="w-full px-4 py-2 bg-gray-900 border border-gray-700 rounded-lg text-white font-mono text-sm"
                  rows="10"
                />
              </div>

              <div className="flex gap-3">
                <button
                  onClick={() => setShowViewModal(false)}
                  className="flex-1 bg-gray-700 hover:bg-gray-600 text-white py-2 rounded-lg transition"
                >
                  Close
                </button>
              </div>
            </div>
          </div>
        </div>
      )}

      {/* Confirm Modal */}
      <ConfirmModal
        isOpen={confirmModal.isOpen}
        onClose={() => setConfirmModal({ ...confirmModal, isOpen: false })}
        onConfirm={confirmModal.onConfirm}
        title={confirmModal.title}
        message={confirmModal.message}
        type={confirmModal.type}
      />
    </div>
  );
};

export default Certificates;
