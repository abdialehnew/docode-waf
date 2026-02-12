
import { useState } from 'react'
import { FileText, Download, Calendar, File } from 'lucide-react'
import api from '../services/api'
import toast from 'react-hot-toast'

const Reports = () => {
    const [loading, setLoading] = useState(false)
    const [dateRange, setDateRange] = useState('7d')
    const [lang, setLang] = useState('en')
    const [customDates, setCustomDates] = useState({
        start: new Date(Date.now() - 7 * 24 * 60 * 60 * 1000).toISOString().split('T')[0],
        end: new Date().toISOString().split('T')[0]
    })

    const handleDownload = async (format) => {
        setLoading(true)
        try {
            let params = { format, lang }

            if (dateRange === 'custom') {
                params.start = customDates.start
                params.end = customDates.end
            } else {
                // Calculate dates based on range
                const end = new Date()
                const start = new Date()

                if (dateRange === '24h') start.setDate(end.getDate() - 1)
                if (dateRange === '7d') start.setDate(end.getDate() - 7)
                if (dateRange === '30d') start.setDate(end.getDate() - 30)

                params.start = start.toISOString().split('T')[0]
                params.end = end.toISOString().split('T')[0]
            }

            const response = await api.get('/reports', {
                params,
                responseType: 'blob'
            })

            // Create download link
            const url = window.URL.createObjectURL(new Blob([response.data]))
            const link = document.createElement('a')
            link.href = url

            const contentDisposition = response.headers['content-disposition']
            let filename = `waf-report-${format}.${format}`
            if (contentDisposition) {
                const fileNameMatch = contentDisposition.match(/filename="?(.+)"?/)
                if (fileNameMatch.length === 2) filename = fileNameMatch[1]
            }

            link.setAttribute('download', filename)
            document.body.appendChild(link)
            link.click()
            link.className.remove() // remove it
            window.URL.revokeObjectURL(url)

            toast.success(`${format.toUpperCase()} Report downloaded successfully`)
        } catch (error) {
            console.error('Download error:', error)
            toast.error(`Failed to download report: ${error.response?.data?.error || error.message}`)
        } finally {
            setLoading(false)
        }
    }

    return (
        <div className="space-y-6">
            <div className="flex justify-between items-center">
                <div>
                    <h1 className="text-2xl font-bold mb-2">Security Reports</h1>
                    <p className="text-gray-600">Generate and download security reports for your WAF.</p>
                </div>
            </div>

            <div className="bg-white p-6 rounded-xl shadow-sm border border-gray-100">
                <h2 className="text-lg font-semibold mb-4 flex items-center gap-2">
                    <Calendar className="w-5 h-5 text-gray-500" />
                    Report Settings
                </h2>

                <div className="space-y-6">
                    <div>
                        <label className="block text-sm font-medium text-gray-700 mb-3">Time Range</label>
                        <div className="flex flex-wrap gap-3">
                            {[
                                { value: '24h', label: 'Last 24 Hours' },
                                { value: '7d', label: 'Last 7 Days' },
                                { value: '30d', label: 'Last 30 Days' },
                                { value: 'custom', label: 'Custom Range' }
                            ].map((option) => (
                                <button
                                    key={option.value}
                                    onClick={() => setDateRange(option.value)}
                                    className={`px-4 py-2 rounded-lg text-sm font-medium transition-all ${dateRange === option.value
                                        ? 'bg-blue-600 text-white shadow-md'
                                        : 'bg-white text-gray-600 border border-gray-200 hover:border-blue-300 hover:bg-blue-50'
                                        }`}
                                >
                                    {option.label}
                                </button>
                            ))}
                        </div>
                    </div>

                    {dateRange === 'custom' && (
                        <div className="bg-gray-50 p-4 rounded-lg border border-gray-200 animate-in fade-in slide-in-from-top-2 duration-200">
                            <div className="grid grid-cols-1 md:grid-cols-2 gap-6">
                                <div>
                                    <label className="block text-sm font-medium text-gray-700 mb-2">Start Date</label>
                                    <div className="relative">
                                        <div className="absolute inset-y-0 left-0 pl-3 flex items-center pointer-events-none">
                                            <Calendar className="h-5 w-5 text-gray-400" />
                                        </div>
                                        <input
                                            type="date"
                                            value={customDates.start}
                                            onChange={(e) => setCustomDates({ ...customDates, start: e.target.value })}
                                            className="pl-10 w-full rounded-lg border-gray-300 focus:ring-blue-500 focus:border-blue-500 shadow-sm"
                                        />
                                    </div>
                                </div>
                                <div>
                                    <label className="block text-sm font-medium text-gray-700 mb-2">End Date</label>
                                    <div className="relative">
                                        <div className="absolute inset-y-0 left-0 pl-3 flex items-center pointer-events-none">
                                            <Calendar className="h-5 w-5 text-gray-400" />
                                        </div>
                                        <input
                                            type="date"
                                            value={customDates.end}
                                            onChange={(e) => setCustomDates({ ...customDates, end: e.target.value })}
                                            className="pl-10 w-full rounded-lg border-gray-300 focus:ring-blue-500 focus:border-blue-500 shadow-sm"
                                        />
                                    </div>
                                </div>
                            </div>
                        </div>
                    )}
                </div>

                <div className="mt-8 grid grid-cols-1 md:grid-cols-2 gap-6 items-end">
                    <div>
                        <label className="block text-sm font-medium text-gray-700 mb-3">Export Format</label>
                        <div className="flex flex-wrap gap-4">
                            <button
                                onClick={() => handleDownload('pdf')}
                                disabled={loading}
                                className="flex items-center gap-3 px-6 py-4 bg-red-50 border border-red-100 rounded-xl hover:bg-red-100 transition-colors group"
                            >
                                <div className="p-2 bg-red-100 rounded-lg group-hover:bg-red-200 transition-colors">
                                    <FileText className="w-6 h-6 text-red-600" />
                                </div>
                                <div className="text-left">
                                    <span className="block font-semibold text-gray-900">PDF Report</span>
                                    <span className="text-sm text-gray-500">Portable Document Format</span>
                                </div>
                                <Download className="w-5 h-5 text-gray-400 ml-2" />
                            </button>

                            <button
                                onClick={() => handleDownload('docx')}
                                disabled={loading}
                                className="flex items-center gap-3 px-6 py-4 bg-blue-50 border border-blue-100 rounded-xl hover:bg-blue-100 transition-colors group"
                            >
                                <div className="p-2 bg-blue-100 rounded-lg group-hover:bg-blue-200 transition-colors">
                                    <File className="w-6 h-6 text-blue-600" />
                                </div>
                                <div className="text-left">
                                    <span className="block font-semibold text-gray-900">DOCX Report</span>
                                    <span className="text-sm text-gray-500">Microsoft Word Document</span>
                                </div>
                                <Download className="w-5 h-5 text-gray-400 ml-2" />
                            </button>
                        </div>
                    </div>

                    <div>
                        <label className="block text-sm font-medium text-gray-700 mb-3">Report Language</label>
                        <div className="flex bg-gray-100 p-1 rounded-lg w-fit">
                            <button
                                onClick={() => setLang('en')}
                                className={`px-4 py-2 rounded-md text-sm font-medium transition-all ${lang === 'en' ? 'bg-white text-blue-600 shadow-sm' : 'text-gray-500 hover:text-gray-700'
                                    }`}
                            >
                                English
                            </button>
                            <button
                                onClick={() => setLang('id')}
                                className={`px-4 py-2 rounded-md text-sm font-medium transition-all ${lang === 'id' ? 'bg-white text-blue-600 shadow-sm' : 'text-gray-500 hover:text-gray-700'
                                    }`}
                            >
                                Indonesia
                            </button>
                        </div>
                    </div>
                </div>

                {loading && (
                    <div className="mt-4 text-center text-gray-500 text-sm animate-pulse">
                        Generating report... This may take a few seconds.
                    </div>
                )}
            </div>

            <div className="grid grid-cols-1 md:grid-cols-3 gap-6">
                {/* Information cards explaining what's in the report */}
                <div className="bg-white p-5 rounded-lg shadow-sm border border-gray-100">
                    <h3 className="font-semibold mb-2">Attack Summary</h3>
                    <p className="text-sm text-gray-500">Detailed breakdown of total requests, blocked attacks, and overall traffic health metrics.</p>
                </div>
                <div className="bg-white p-5 rounded-lg shadow-sm border border-gray-100">
                    <h3 className="font-semibold mb-2">Top Threats</h3>
                    <p className="text-sm text-gray-500">List of top attacking IP addresses and countries of origin for the selected period.</p>
                </div>
                <div className="bg-white p-5 rounded-lg shadow-sm border border-gray-100">
                    <h3 className="font-semibold mb-2">Recent Logs</h3>
                    <p className="text-sm text-gray-500">A snapshot of the most recent blocked access attempts with timestamps and attack types.</p>
                </div>
            </div>
        </div>
    )
}

export default Reports
