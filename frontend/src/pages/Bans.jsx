import { useEffect, useState } from 'react'
import { getBans, unbanIP } from '../services/api'
import { Trash2, AlertTriangle, Shield, Clock } from 'lucide-react'
import ConfirmModal from '../components/ConfirmModal'
import logger from '../utils/logger'

const Bans = () => {
    const [bans, setBans] = useState([])
    const [loading, setLoading] = useState(true)
    const [confirmModal, setConfirmModal] = useState({ isOpen: false, onConfirm: null, title: '', message: '' })

    useEffect(() => {
        loadBans()
    }, [])

    const loadBans = async () => {
        try {
            const response = await getBans()
            setBans(response.data || [])
        } catch (error) {
            logger.error('Failed to load bans:', error)
            setBans([])
        } finally {
            setLoading(false)
        }
    }

    const handleUnban = (id, ip) => {
        setConfirmModal({
            isOpen: true,
            title: 'Unban IP',
            message: `Are you sure you want to unban ${ip}? This will allow immediate access.`,
            type: 'warning',
            onConfirm: async () => {
                try {
                    await unbanIP(id)
                    loadBans()
                } catch (error) {
                    logger.error('Failed to unban IP:', error)
                    // Ideally show a toast notification here
                }
            }
        })
    }

    if (loading) {
        return <div className="text-center py-12">Loading...</div>
    }

    return (
        <div>
            <div className="flex justify-between items-center mb-8">
                <div>
                    <h1 className="text-3xl font-bold flex items-center gap-2">
                        <Shield className="w-8 h-8 text-red-600" />
                        Dynamic IP Bans
                    </h1>
                    <p className="text-gray-500 mt-1">
                        IP addresses temporarily banned due to suspicious activity (Fail2Ban style).
                    </p>
                </div>
                <button
                    onClick={loadBans}
                    className="btn btn-secondary"
                >
                    Refresh
                </button>
            </div>

            <div className="card overflow-hidden p-0">
                <div className="overflow-x-auto">
                    <table className="w-full">
                        <thead className="bg-gray-50 border-b">
                            <tr>
                                <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
                                    IP Address
                                </th>
                                <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
                                    Reason
                                </th>
                                <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
                                    Violations
                                </th>
                                <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
                                    Banned At
                                </th>
                                <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
                                    Expires At
                                </th>
                                <th className="px-6 py-3 text-right text-xs font-medium text-gray-500 uppercase tracking-wider">
                                    Actions
                                </th>
                            </tr>
                        </thead>
                        <tbody className="divide-y divide-gray-200">
                            {bans.length === 0 ? (
                                <tr>
                                    <td colSpan="6" className="px-6 py-8 text-center text-gray-500">
                                        <div className="flex flex-col items-center gap-2">
                                            <Shield className="w-12 h-12 text-gray-300" />
                                            <p>No active bans found.</p>
                                            <p className="text-xs text-gray-400">Everything looks safe!</p>
                                        </div>
                                    </td>
                                </tr>
                            ) : (
                                bans.map((ban) => (
                                    <tr key={ban.id} className="hover:bg-gray-50">
                                        <td className="px-6 py-4 whitespace-nowrap">
                                            <div className="font-mono text-sm font-medium text-gray-900">
                                                {ban.ip_address}
                                            </div>
                                        </td>
                                        <td className="px-6 py-4">
                                            <div className="text-sm text-gray-500 max-w-xs truncate" title={ban.reason}>
                                                {ban.reason}
                                            </div>
                                        </td>
                                        <td className="px-6 py-4 whitespace-nowrap">
                                            <span className="px-2 inline-flex text-xs leading-5 font-semibold rounded-full bg-red-100 text-red-800">
                                                {ban.violation_count}
                                            </span>
                                        </td>
                                        <td className="px-6 py-4 whitespace-nowrap text-sm text-gray-500">
                                            {new Date(ban.banned_at).toLocaleString()}
                                        </td>
                                        <td className="px-6 py-4 whitespace-nowrap text-sm text-gray-500">
                                            <div className="flex items-center gap-1">
                                                <Clock className="w-3 h-3" />
                                                {new Date(ban.expires_at).toLocaleString()}
                                            </div>
                                        </td>
                                        <td className="px-6 py-4 whitespace-nowrap text-right text-sm font-medium">
                                            <button
                                                onClick={() => handleUnban(ban.id, ban.ip_address)}
                                                className="text-red-600 hover:text-red-900 flex items-center gap-1 ml-auto"
                                                title="Unban IP"
                                            >
                                                <Trash2 className="w-4 h-4" />
                                                Unban
                                            </button>
                                        </td>
                                    </tr>
                                ))
                            )}
                        </tbody>
                    </table>
                </div>
            </div>

            <ConfirmModal
                isOpen={confirmModal.isOpen}
                onClose={() => setConfirmModal({ ...confirmModal, isOpen: false })}
                onConfirm={() => {
                    confirmModal.onConfirm()
                    setConfirmModal({ ...confirmModal, isOpen: false })
                }}
                title={confirmModal.title}
                message={confirmModal.message}
                type={confirmModal.type}
            />
        </div>
    )
}

export default Bans
