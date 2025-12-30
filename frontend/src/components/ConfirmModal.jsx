import React from 'react';
import { AlertTriangle, CheckCircle, Info, X } from 'lucide-react';

const ConfirmModal = ({ 
  isOpen, 
  onClose, 
  onConfirm, 
  title = 'Confirmation',
  message = 'Are you sure?',
  confirmText = 'Confirm',
  cancelText = 'Cancel',
  type = 'warning' // 'warning', 'danger', 'info', 'success'
}) => {
  if (!isOpen) return null;

  const typeConfig = {
    warning: {
      icon: AlertTriangle,
      iconClass: 'text-yellow-500',
      buttonClass: 'bg-yellow-600 hover:bg-yellow-700',
    },
    danger: {
      icon: AlertTriangle,
      iconClass: 'text-red-500',
      buttonClass: 'bg-red-600 hover:bg-red-700',
    },
    info: {
      icon: Info,
      iconClass: 'text-blue-500',
      buttonClass: 'bg-blue-600 hover:bg-blue-700',
    },
    success: {
      icon: CheckCircle,
      iconClass: 'text-green-500',
      buttonClass: 'bg-green-600 hover:bg-green-700',
    }
  };

  const config = typeConfig[type] || typeConfig.warning;
  const Icon = config.icon;

  const handleConfirm = () => {
    onConfirm();
    onClose();
  };

  return (
    <div 
      className="fixed inset-0 bg-black bg-opacity-50 flex items-center justify-center z-50"
      role="dialog"
      aria-modal="true"
      aria-labelledby="modal-title"
    >
      <div className="bg-white rounded-lg shadow-xl max-w-md w-full mx-4">
        {/* Header */}
        <div className="flex items-center justify-between p-6 border-b border-gray-200">
          <div className="flex items-center space-x-3">
            <Icon className={`w-6 h-6 ${config.iconClass}`} />
            <h3 id="modal-title" className="text-lg font-semibold text-gray-900">
              {title}
            </h3>
          </div>
          <button
            onClick={onClose}
            className="text-gray-400 hover:text-gray-600 transition-colors"
            aria-label="Close modal"
          >
            <X className="w-5 h-5" />
          </button>
        </div>

        {/* Body */}
        <div className="p-6">
          <p className="text-gray-600">{message}</p>
        </div>

        {/* Footer */}
        <div className="flex justify-end space-x-3 p-6 bg-gray-50 rounded-b-lg">
          <button
            onClick={onClose}
            className="px-4 py-2 border border-gray-300 rounded-lg text-gray-700 hover:bg-gray-100 transition-colors"
          >
            {cancelText}
          </button>
          <button
            onClick={handleConfirm}
            className={`px-4 py-2 text-white rounded-lg transition-colors ${config.buttonClass}`}
          >
            {confirmText}
          </button>
        </div>
      </div>
    </div>
  );
};

export default ConfirmModal;
