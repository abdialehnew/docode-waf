import React, { useState, useEffect } from 'react';
import { Card, CardContent, CardHeader, CardTitle } from '../components/ui/card';
import { Tabs, TabsContent, TabsList, TabsTrigger } from '../components/ui/tabs';
import api from '../services/api';

const Monitoring = () => {
  const [activeTab, setActiveTab] = useState('nginx');
  const [vhosts, setVhosts] = useState([]);
  const [selectedVhost, setSelectedVhost] = useState('');
  const [logType, setLogType] = useState('access');
  const [nginxLogs, setNginxLogs] = useState([]);
  const [wafLogs, setWafLogs] = useState([]);
  const [loading, setLoading] = useState(false);
  const [liveMode, setLiveMode] = useState(false);
  const [dateRange, setDateRange] = useState({
    start: new Date(Date.now() - 24 * 60 * 60 * 1000).toISOString().split('T')[0],
    end: new Date().toISOString().split('T')[0]
  });

  useEffect(() => {
    fetchVhosts();
  }, []);

  useEffect(() => {
    if (selectedVhost && activeTab === 'nginx') {
      fetchNginxLogs();
    }
  }, [selectedVhost, logType]);

  useEffect(() => {
    if (activeTab === 'waf') {
      fetchWAFLogs();
    }
  }, [activeTab, dateRange]);

  useEffect(() => {
    let interval;
    if (liveMode && activeTab === 'nginx' && selectedVhost) {
      interval = setInterval(() => {
        fetchNginxLogs();
      }, 3000);
    } else if (liveMode && activeTab === 'waf') {
      interval = setInterval(() => {
        fetchWAFLogs();
      }, 3000);
    }
    return () => {
      if (interval) clearInterval(interval);
    };
  }, [liveMode, activeTab, selectedVhost]);

  const fetchVhosts = async () => {
    try {
      const response = await api.get('/logs/vhosts');
      setVhosts(response.data);
      if (response.data.length > 0) {
        setSelectedVhost(response.data[0].domain);
      }
    } catch (error) {
      console.error('Failed to fetch vhosts:', error);
    }
  };

  const fetchNginxLogs = async () => {
    if (!selectedVhost) return;
    
    setLoading(true);
    try {
      const endpoint = logType === 'access' ? '/logs/nginx/access' : '/logs/nginx/error';
      const response = await api.get(endpoint, {
        params: { domain: selectedVhost, lines: 100 }
      });
      setNginxLogs(response.data.lines || []);
    } catch (error) {
      console.error('Failed to fetch nginx logs:', error);
    } finally {
      setLoading(false);
    }
  };

  const fetchWAFLogs = async () => {
    setLoading(true);
    try {
      const response = await api.get('/logs/waf', {
        params: {
          start_date: dateRange.start,
          end_date: dateRange.end,
          limit: 100
        }
      });
      setWafLogs(response.data.logs || []);
    } catch (error) {
      console.error('Failed to fetch WAF logs:', error);
    } finally {
      setLoading(false);
    }
  };

  const formatTimestamp = (timestamp) => {
    return new Date(timestamp).toLocaleString('en-US', {
      year: 'numeric',
      month: 'short',
      day: '2-digit',
      hour: '2-digit',
      minute: '2-digit',
      second: '2-digit'
    });
  };

  return (
    <div className="space-y-6">
      <div className="flex items-center justify-between">
        <h1 className="text-3xl font-bold">Monitoring & Logs</h1>
        <div className="flex items-center gap-2">
          <label className="flex items-center gap-2 cursor-pointer">
            <input
              type="checkbox"
              checked={liveMode}
              onChange={(e) => setLiveMode(e.target.checked)}
              className="w-4 h-4"
            />
            <span className="text-sm font-medium">Live Mode</span>
          </label>
          {liveMode && (
            <span className="flex items-center gap-1 text-sm text-green-500">
              <span className="w-2 h-2 bg-green-500 rounded-full animate-pulse"></span>
              Live
            </span>
          )}
        </div>
      </div>

      <Tabs value={activeTab} onValueChange={setActiveTab}>
        <TabsList className="grid w-full grid-cols-2">
          <TabsTrigger value="nginx">Nginx Logs</TabsTrigger>
          <TabsTrigger value="waf">WAF Logs</TabsTrigger>
        </TabsList>

        <TabsContent value="nginx" className="space-y-4">
          <Card>
            <CardHeader>
              <CardTitle>Nginx Logs</CardTitle>
            </CardHeader>
            <CardContent className="space-y-4">
              {/* Filters */}
              <div className="flex gap-4">
                <div className="flex-1">
                  <label className="block text-sm font-medium mb-2">Virtual Host</label>
                  <select
                    value={selectedVhost}
                    onChange={(e) => setSelectedVhost(e.target.value)}
                    className="w-full px-3 py-2 border rounded-lg focus:ring-2 focus:ring-blue-500"
                  >
                    {vhosts.map((vhost) => (
                      <option key={vhost.domain} value={vhost.domain}>
                        {vhost.name} ({vhost.domain})
                      </option>
                    ))}
                  </select>
                </div>

                <div className="flex-1">
                  <label className="block text-sm font-medium mb-2">Log Type</label>
                  <select
                    value={logType}
                    onChange={(e) => setLogType(e.target.value)}
                    className="w-full px-3 py-2 border rounded-lg focus:ring-2 focus:ring-blue-500"
                  >
                    <option value="access">Access Logs</option>
                    <option value="error">Error Logs</option>
                  </select>
                </div>
              </div>

              {/* Logs Display */}
              <div className="bg-gray-900 text-gray-100 p-4 rounded-lg font-mono text-sm overflow-auto max-h-[600px]">
                {loading ? (
                  <div className="text-center text-gray-400">Loading logs...</div>
                ) : nginxLogs.length === 0 ? (
                  <div className="text-center text-gray-400">No logs available</div>
                ) : (
                  nginxLogs.map((log, index) => (
                    <div key={index} className="py-1 hover:bg-gray-800">
                      {log}
                    </div>
                  ))
                )}
              </div>

              <div className="flex justify-between items-center text-sm text-gray-500">
                <span>Showing last {nginxLogs.length} lines</span>
                <button
                  onClick={fetchNginxLogs}
                  className="px-4 py-2 bg-blue-500 text-white rounded-lg hover:bg-blue-600"
                >
                  Refresh
                </button>
              </div>
            </CardContent>
          </Card>
        </TabsContent>

        <TabsContent value="waf" className="space-y-4">
          <Card>
            <CardHeader>
              <CardTitle>WAF Logs</CardTitle>
            </CardHeader>
            <CardContent className="space-y-4">
              {/* Date Range Filter */}
              <div className="flex gap-4">
                <div className="flex-1">
                  <label className="block text-sm font-medium mb-2">Start Date</label>
                  <input
                    type="date"
                    value={dateRange.start}
                    onChange={(e) => setDateRange({ ...dateRange, start: e.target.value })}
                    className="w-full px-3 py-2 border rounded-lg focus:ring-2 focus:ring-blue-500"
                  />
                </div>
                <div className="flex-1">
                  <label className="block text-sm font-medium mb-2">End Date</label>
                  <input
                    type="date"
                    value={dateRange.end}
                    onChange={(e) => setDateRange({ ...dateRange, end: e.target.value })}
                    className="w-full px-3 py-2 border rounded-lg focus:ring-2 focus:ring-blue-500"
                  />
                </div>
                <div className="flex items-end">
                  <button
                    onClick={fetchWAFLogs}
                    className="px-4 py-2 bg-blue-500 text-white rounded-lg hover:bg-blue-600"
                  >
                    Apply
                  </button>
                </div>
              </div>

              {/* WAF Logs Table */}
              <div className="overflow-x-auto">
                <table className="w-full text-sm">
                  <thead className="bg-gray-50">
                    <tr>
                      <th className="px-4 py-3 text-left">Timestamp</th>
                      <th className="px-4 py-3 text-left">Client IP</th>
                      <th className="px-4 py-3 text-left">Method</th>
                      <th className="px-4 py-3 text-left">URL</th>
                      <th className="px-4 py-3 text-left">Status</th>
                      <th className="px-4 py-3 text-left">Attack</th>
                      <th className="px-4 py-3 text-left">Blocked</th>
                    </tr>
                  </thead>
                  <tbody className="divide-y">
                    {loading ? (
                      <tr>
                        <td colSpan="7" className="px-4 py-8 text-center text-gray-400">
                          Loading logs...
                        </td>
                      </tr>
                    ) : wafLogs.length === 0 ? (
                      <tr>
                        <td colSpan="7" className="px-4 py-8 text-center text-gray-400">
                          No logs found
                        </td>
                      </tr>
                    ) : (
                      wafLogs.map((log) => (
                        <tr key={log.id} className="hover:bg-gray-50">
                          <td className="px-4 py-3 whitespace-nowrap">
                            {formatTimestamp(log.timestamp)}
                          </td>
                          <td className="px-4 py-3">
                            {log.client_ip}
                            {log.country_code && (
                              <span className="ml-2 text-xs text-gray-500">
                                {log.country_code}
                              </span>
                            )}
                          </td>
                          <td className="px-4 py-3">{log.method}</td>
                          <td className="px-4 py-3 max-w-xs truncate" title={log.url}>
                            {log.url}
                          </td>
                          <td className="px-4 py-3">
                            <span className={`px-2 py-1 rounded text-xs ${
                              log.status_code >= 500 ? 'bg-red-100 text-red-800' :
                              log.status_code >= 400 ? 'bg-yellow-100 text-yellow-800' :
                              log.status_code >= 300 ? 'bg-blue-100 text-blue-800' :
                              'bg-green-100 text-green-800'
                            }`}>
                              {log.status_code}
                            </span>
                          </td>
                          <td className="px-4 py-3">
                            {log.is_attack ? (
                              <span className="px-2 py-1 bg-red-100 text-red-800 rounded text-xs">
                                {log.attack_type || 'Attack'}
                              </span>
                            ) : (
                              <span className="text-gray-400">-</span>
                            )}
                          </td>
                          <td className="px-4 py-3">
                            {log.blocked ? (
                              <span className="px-2 py-1 bg-red-100 text-red-800 rounded text-xs">
                                Blocked
                              </span>
                            ) : (
                              <span className="px-2 py-1 bg-green-100 text-green-800 rounded text-xs">
                                Allowed
                              </span>
                            )}
                          </td>
                        </tr>
                      ))
                    )}
                  </tbody>
                </table>
              </div>

              <div className="flex justify-between items-center text-sm text-gray-500">
                <span>Total: {wafLogs.length} records</span>
                <span>
                  {dateRange.start} to {dateRange.end}
                </span>
              </div>
            </CardContent>
          </Card>
        </TabsContent>
      </Tabs>
    </div>
  );
};

export default Monitoring;
