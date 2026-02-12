import React, { useState, useEffect } from 'react'
import api from '../services/api'
import Swal from 'sweetalert2'
import {
    Plus,
    Trash2,
    Edit,
    CheckCircle,
    XCircle,
    Shield,
    AlertTriangle,
    Eye,
    ArrowUp,
    ArrowDown
} from 'lucide-react'

const Rules = () => {
    const [rules, setRules] = useState([])
    const [loading, setLoading] = useState(true)
    const [showModal, setShowModal] = useState(false)
    const [editingRule, setEditingRule] = useState(null)

    // Form State
    const [formData, setFormData] = useState({
        name: '',
        description: '',
        priority: 10,
        action: 'block',
        match_logic: 'AND',
        enabled: true,
        conditions: []
    })

    useEffect(() => {
        fetchRules()
    }, [])

    const fetchRules = async () => {
        try {
            setLoading(true)
            const response = await api.get('/rules')
            setRules(response.data || [])
        } catch (error) {
            console.error('Error fetching rules:', error)
            Swal.fire('Error', 'Failed to fetch rules', 'error')
        } finally {
            setLoading(false)
        }
    }

    const resetForm = () => {
        setFormData({
            name: '',
            description: '',
            priority: 10,
            action: 'block',
            match_logic: 'AND',
            enabled: true,
            conditions: []
        })
        setEditingRule(null)
    }

    const handleEdit = (rule) => {
        // Ensure conditions is an array (backend might return null)
        let conditions = rule.conditions
        if (typeof conditions === 'string') {
            try { conditions = JSON.parse(conditions) } catch (e) { }
        }

        setFormData({
            name: rule.name,
            description: rule.description || '',
            priority: rule.priority,
            action: rule.action,
            match_logic: rule.match_logic || 'AND',
            enabled: rule.enabled,
            conditions: Array.isArray(conditions) ? conditions : []
        })
        setEditingRule(rule)
        setShowModal(true)
    }

    const handleDelete = async (id) => {
        const result = await Swal.fire({
            title: 'Are you sure?',
            text: "You won't be able to revert this!",
            icon: 'warning',
            showCancelButton: true,
            confirmButtonText: 'Yes, delete it!'
        })

        if (result.isConfirmed) {
            try {
                await api.delete(`/rules/${id}`)
                fetchRules()
                Swal.fire('Deleted!', 'Rule has been deleted.', 'success')
            } catch (error) {
                Swal.fire('Error', 'Failed to delete rule', 'error')
            }
        }
    }

    const handleSubmit = async (e) => {
        e.preventDefault()
        try {
            const payload = {
                ...formData,
                priority: parseInt(formData.priority)
            }

            if (editingRule) {
                await api.put(`/rules/${editingRule.id}`, payload)
            } else {
                await api.post('/rules', payload)
            }

            setShowModal(false)
            fetchRules()
            resetForm()
            Swal.fire('Success', 'Rule saved successfully', 'success')
        } catch (error) {
            console.error(error)
            Swal.fire('Error', 'Failed to save rule', 'error')
        }
    }

    // Condition Editor Helpers
    const addCondition = () => {
        setFormData({
            ...formData,
            conditions: [...formData.conditions, { field: 'ip', operator: 'eq', value: '' }]
        })
    }

    const removeCondition = (index) => {
        const newConditions = [...formData.conditions]
        newConditions.splice(index, 1)
        setFormData({ ...formData, conditions: newConditions })
    }

    const updateCondition = (index, key, value) => {
        const newConditions = [...formData.conditions]
        newConditions[index][key] = value
        setFormData({ ...formData, conditions: newConditions })
    }

    const getActionIcon = (action) => {
        switch (action) {
            case 'block': return <Shield className="w-4 h-4 text-red-500" />
            case 'allow': return <CheckCircle className="w-4 h-4 text-green-500" />
            case 'challenge': return <AlertTriangle className="w-4 h-4 text-yellow-500" />
            case 'log': return <Eye className="w-4 h-4 text-blue-500" />
            default: return null
        }
    }

    const getActionBadgeColor = (action) => {
        switch (action) {
            case 'block': return 'bg-red-900/30 text-red-500 border-red-800'
            case 'allow': return 'bg-green-900/30 text-green-500 border-green-800'
            case 'challenge': return 'bg-yellow-900/30 text-yellow-500 border-yellow-800'
            case 'log': return 'bg-blue-900/30 text-blue-500 border-blue-800'
            default: return 'bg-gray-800 text-gray-400'
        }
    }

    return (
        <div className="space-y-6">
            <div className="flex justify-between items-center">
                <div>
                    <h1 className="text-2xl font-bold text-white">Custom Rules</h1>
                    <p className="text-gray-400">Manage visual security rules logic</p>
                </div>
                <button
                    onClick={() => { resetForm(); setShowModal(true) }}
                    className="px-4 py-2 bg-blue-600 hover:bg-blue-700 text-white rounded-lg flex items-center gap-2"
                >
                    <Plus size={18} /> Add Rule
                </button>
            </div>

            {/* Rules List */}
            <div className="bg-gray-900 border border-gray-800 rounded-xl overflow-hidden">
                {loading ? (
                    <div className="p-8 text-center text-gray-400">Loading rules...</div>
                ) : rules.length === 0 ? (
                    <div className="p-8 text-center text-gray-400">No rules found. Create one to get started.</div>
                ) : (
                    <div className="divide-y divide-gray-800">
                        {rules.map((rule) => (
                            <div key={rule.id} className="p-4 flex items-center justify-between hover:bg-gray-800/50 transition-colors">
                                <div className="flex items-center gap-4">
                                    <div className="flex flex-col items-center justify-center w-10 h-10 bg-gray-800 rounded-lg text-gray-400 font-mono text-xs">
                                        <span className="text-[10px] uppercase">PRI</span>
                                        <span className="font-bold text-white">{rule.priority}</span>
                                    </div>
                                    <div>
                                        <div className="flex items-center gap-2">
                                            <h3 className="font-medium text-white">{rule.name}</h3>
                                            {!rule.enabled && (
                                                <span className="text-xs px-2 py-0.5 rounded-full bg-gray-800 text-gray-500">Disabled</span>
                                            )}
                                        </div>
                                        <p className="text-sm text-gray-500">{rule.description || 'No description'}</p>
                                        <div className="flex items-center gap-2 mt-1">
                                            <span className={`text-xs px-2 py-0.5 rounded border flex items-center gap-1 ${getActionBadgeColor(rule.action)}`}>
                                                {getActionIcon(rule.action)}
                                                {rule.action.toUpperCase()}
                                            </span>
                                            <span className="text-xs text-gray-500">
                                                {Array.isArray(rule.conditions) ? rule.conditions.length : 0} conditions
                                            </span>
                                        </div>
                                    </div>
                                </div>
                                <div className="flex items-center gap-2">
                                    <button
                                        onClick={() => handleEdit(rule)}
                                        className="p-2 text-gray-400 hover:text-white hover:bg-gray-700 rounded-lg"
                                    >
                                        <Edit size={18} />
                                    </button>
                                    <button
                                        onClick={() => handleDelete(rule.id)}
                                        className="p-2 text-red-400 hover:text-red-300 hover:bg-red-900/20 rounded-lg"
                                    >
                                        <Trash2 size={18} />
                                    </button>
                                </div>
                            </div>
                        ))}
                    </div>
                )}
            </div>

            {/* Modal */}
            {showModal && (
                <div className="fixed inset-0 z-50 flex items-center justify-center bg-black/50 backdrop-blur-sm p-4">
                    <div className="bg-gray-900 border border-gray-800 rounded-xl w-full max-w-2xl max-h-[90vh] overflow-y-auto shadow-2xl">
                        <div className="p-6 border-b border-gray-800 flex justify-between items-center sticky top-0 bg-gray-900 z-10">
                            <h2 className="text-xl font-bold text-white">
                                {editingRule ? 'Edit Rule' : 'New Rule'}
                            </h2>
                            <button onClick={() => setShowModal(false)} className="text-gray-400 hover:text-white">
                                <XCircle size={24} />
                            </button>
                        </div>

                        <form onSubmit={handleSubmit} className="p-6 space-y-6">
                            {/* Basic Info */}
                            <div className="grid grid-cols-2 gap-4">
                                <div className="space-y-1">
                                    <label className="text-sm text-gray-400">Rule Name</label>
                                    <input
                                        type="text"
                                        required
                                        className="w-full bg-gray-800 border border-gray-700 rounded-lg px-3 py-2 text-white focus:outline-none focus:border-blue-500"
                                        value={formData.name}
                                        onChange={e => setFormData({ ...formData, name: e.target.value })}
                                        placeholder="e.g., Block Admin Access"
                                    />
                                </div>
                                <div className="space-y-1">
                                    <label className="text-sm text-gray-400">Priority (Higher runs first)</label>
                                    <input
                                        type="number"
                                        required
                                        className="w-full bg-gray-800 border border-gray-700 rounded-lg px-3 py-2 text-white focus:outline-none focus:border-blue-500"
                                        value={formData.priority}
                                        onChange={e => setFormData({ ...formData, priority: e.target.value })}
                                    />
                                </div>
                            </div>

                            <div className="space-y-1">
                                <label className="text-sm text-gray-400">Description</label>
                                <textarea
                                    className="w-full bg-gray-800 border border-gray-700 rounded-lg px-3 py-2 text-white focus:outline-none focus:border-blue-500 h-20 resize-none"
                                    value={formData.description}
                                    onChange={e => setFormData({ ...formData, description: e.target.value })}
                                    placeholder="Explain what this rule does..."
                                />
                            </div>

                            <div className="grid grid-cols-2 gap-4">
                                <div className="space-y-1">
                                    <label className="text-sm text-gray-400">Action</label>
                                    <select
                                        className="w-full bg-gray-800 border border-gray-700 rounded-lg px-3 py-2 text-white focus:outline-none focus:border-blue-500"
                                        value={formData.action}
                                        onChange={e => setFormData({ ...formData, action: e.target.value })}
                                    >
                                        <option value="block">Block (403)</option>
                                        <option value="allow">Allow (Bypass WAF)</option>
                                        <option value="challenge">Challenge (Captcha)</option>
                                        <option value="log">Log Only</option>
                                    </select>
                                </div>
                                <div className="flex items-center pt-6">
                                    <label className="flex items-center gap-2 cursor-pointer">
                                        <input
                                            type="checkbox"
                                            className="w-5 h-5 rounded border-gray-700 bg-gray-800 text-blue-600 focus:ring-0"
                                            checked={formData.enabled}
                                            onChange={e => setFormData({ ...formData, enabled: e.target.checked })}
                                        />
                                        <span className="text-white">Enable Rule</span>
                                    </label>
                                </div>
                            </div>

                            <div className="space-y-1">
                                <label className="text-sm text-gray-400">Condition Logic</label>
                                <div className="flex gap-4">
                                    <label className="flex items-center gap-2 cursor-pointer">
                                        <input
                                            type="radio"
                                            name="match_logic"
                                            value="AND"
                                            checked={!formData.match_logic || formData.match_logic === 'AND'}
                                            onChange={e => setFormData({ ...formData, match_logic: e.target.value })}
                                            className="w-4 h-4 text-blue-600 border-gray-700 bg-gray-800 focus:ring-0"
                                        />
                                        <span className="text-white text-sm">Match ALL Conditions (AND)</span>
                                    </label>
                                    <label className="flex items-center gap-2 cursor-pointer">
                                        <input
                                            type="radio"
                                            name="match_logic"
                                            value="OR"
                                            checked={formData.match_logic === 'OR'}
                                            onChange={e => setFormData({ ...formData, match_logic: e.target.value })}
                                            className="w-4 h-4 text-blue-600 border-gray-700 bg-gray-800 focus:ring-0"
                                        />
                                        <span className="text-white text-sm">Match ANY Condition (OR)</span>
                                    </label>
                                </div>
                            </div>

                            {/* Conditions Builder */}
                            <div className="space-y-3">
                                <div className="flex justify-between items-center">
                                    <label className="text-sm font-medium text-gray-300">Conditions (AND logic)</label>
                                    <button
                                        type="button"
                                        onClick={addCondition}
                                        className="text-xs px-2 py-1 bg-gray-800 hover:bg-gray-700 text-blue-400 rounded border border-gray-700"
                                    >
                                        + Add Condition
                                    </button>
                                </div>

                                {formData.conditions.length === 0 && (
                                    <div className="p-4 border border-dashed border-gray-700 rounded-lg text-center text-sm text-gray-500">
                                        No conditions. This rule will match ALL requests!
                                    </div>
                                )}

                                {formData.conditions.map((cond, index) => (
                                    <div key={index} className="flex gap-2 items-center bg-gray-800/50 p-2 rounded border border-gray-700">
                                        <select
                                            className="bg-gray-800 border border-gray-700 rounded px-2 py-1 text-sm text-white focus:outline-none w-1/4"
                                            value={cond.field}
                                            onChange={e => updateCondition(index, 'field', e.target.value)}
                                        >
                                            <option value="ip">IP Address</option>
                                            <option value="country">Country Code</option>
                                            <option value="path">URL Path</option>
                                            <option value="method">HTTP Method</option>
                                            <option value="user_agent">User Agent</option>
                                            <option value="header.Referer">Referer</option>
                                            <option value="query.id">Query Param (id)</option>
                                        </select>

                                        <select
                                            className="bg-gray-800 border border-gray-700 rounded px-2 py-1 text-sm text-white focus:outline-none w-1/4"
                                            value={cond.operator}
                                            onChange={e => updateCondition(index, 'operator', e.target.value)}
                                        >
                                            <option value="eq">Equals</option>
                                            <option value="neq">Not Equals</option>
                                            <option value="contains">Contains</option>
                                            <option value="not_contains">Not Contains</option>
                                            <option value="starts_with">Starts With</option>
                                            <option value="ends_with">Ends With</option>
                                            <option value="regex">Regex</option>
                                            <option value="cidr_contains">CIDR Contains</option>
                                        </select>

                                        <input
                                            type="text"
                                            className="bg-gray-800 border border-gray-700 rounded px-2 py-1 text-sm text-white focus:outline-none flex-1"
                                            value={cond.value}
                                            onChange={e => updateCondition(index, 'value', e.target.value)}
                                            placeholder="Value..."
                                        />

                                        <button
                                            type="button"
                                            onClick={() => removeCondition(index)}
                                            className="text-red-400 hover:text-red-300 p-1"
                                        >
                                            <XCircle size={16} />
                                        </button>
                                    </div>
                                ))}
                            </div>

                            <div className="pt-4 border-t border-gray-800 flex justify-end gap-3">
                                <button
                                    type="button"
                                    onClick={() => setShowModal(false)}
                                    className="px-4 py-2 text-gray-400 hover:text-white"
                                >
                                    Cancel
                                </button>
                                <button
                                    type="submit"
                                    className="px-4 py-2 bg-blue-600 hover:bg-blue-700 text-white rounded-lg"
                                >
                                    Save Rule
                                </button>
                            </div>
                        </form>
                    </div>
                </div>
            )}
        </div>
    )
}

export default Rules
