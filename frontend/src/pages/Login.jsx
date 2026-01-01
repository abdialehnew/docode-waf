import React, { useState, useEffect } from 'react';
import { Link, useNavigate } from 'react-router-dom';
import { useAuth } from '../context/AuthContext';
import { Shield, Mail, Lock, AlertCircle, Eye, EyeOff } from 'lucide-react';
import { getAppSettings, getTurnstileSiteKey } from '../services/api';
import Turnstile from '../components/Turnstile';

const Login = () => {
  const navigate = useNavigate();
  const { login } = useAuth();
  const [formData, setFormData] = useState({
    username: '',
    password: ''
  });
  const [error, setError] = useState('');
  const [loading, setLoading] = useState(false);
  const [showPassword, setShowPassword] = useState(false);
  const [appSettings, setAppSettings] = useState({
    app_name: 'Docode WAF',
    app_logo: '',
    signup_enabled: true
  });
  const [turnstileConfig, setTurnstileConfig] = useState({
    enabled: false,
    site_key: ''
  });
  const [turnstileToken, setTurnstileToken] = useState('');

  useEffect(() => {
    loadAppSettings();
    loadTurnstileConfig();
  }, []);

  const loadAppSettings = async () => {
    try {
      const response = await getAppSettings();
      if (response.data) {
        setAppSettings(response.data);
        // Update document title
        document.title = response.data.app_name || 'Docode WAF';
      }
    } catch (error) {
      console.error('Failed to load app settings:', error);
    }
  };

  const loadTurnstileConfig = async () => {
    try {
      const response = await getTurnstileSiteKey();
      if (response.data) {
        setTurnstileConfig(response.data);
      }
    } catch (error) {
      console.error('Failed to load Turnstile config:', error);
    }
  };

  const handleChange = (e) => {
    setFormData({
      ...formData,
      [e.target.name]: e.target.value
    });
    setError('');
  };

  const handleSubmit = async (e) => {
    e.preventDefault();
    setLoading(true);
    setError('');

    // Validate Turnstile if enabled
    if (turnstileConfig.enabled && !turnstileToken) {
      setError('Please complete the captcha verification');
      setLoading(false);
      return;
    }

    try {
      await login(formData.username, formData.password, turnstileToken);
      navigate('/');
    } catch (err) {
      setError(err.response?.data?.error || 'Login failed. Please try again.');
      // Reset Turnstile on error
      setTurnstileToken('');
    } finally {
      setLoading(false);
    }
  };

  return (
    <div className="min-h-screen bg-gradient-to-br from-gray-900 via-gray-800 to-gray-900 flex items-center justify-center p-4">
      <div className="max-w-md w-full">
        {/* Logo/Header */}
        <div className="text-center mb-8">
          {appSettings.app_logo ? (
            <div className="flex justify-center mb-4">
              <img 
                src={appSettings.app_logo} 
                alt={appSettings.app_name} 
                className="h-16 w-auto object-contain"
              />
            </div>
          ) : (
            <div className="inline-flex items-center justify-center w-16 h-16 bg-blue-600 rounded-full mb-4">
              <Shield className="w-8 h-8 text-white" />
            </div>
          )}
          <h1 className="text-3xl font-bold text-white mb-2">{appSettings.app_name}</h1>
          <p className="text-gray-400">Web Application Firewall</p>
        </div>

        {/* Login Form */}
        <div className="bg-gray-800 rounded-lg shadow-xl p-8">
          <h2 className="text-2xl font-bold text-white mb-6">Sign In</h2>

          {error && (
            <div className="mb-4 p-4 bg-red-500/10 border border-red-500 rounded-lg flex items-center gap-2">
              <AlertCircle className="w-5 h-5 text-red-500" />
              <p className="text-red-500 text-sm">{error}</p>
            </div>
          )}

          <form onSubmit={handleSubmit} className="space-y-4">
            <div>
              <label htmlFor="username" className="block text-sm font-medium text-gray-300 mb-2">
                Username or Email
              </label>
              <div className="relative">
                <Mail className="absolute left-3 top-1/2 transform -translate-y-1/2 w-5 h-5 text-gray-400" />
                <input
                  id="username"
                  type="text"
                  name="username"
                  value={formData.username}
                  onChange={handleChange}
                  className="w-full pl-10 pr-4 py-2 bg-gray-700 border border-gray-600 rounded-lg text-white placeholder-gray-400 focus:outline-none focus:ring-2 focus:ring-blue-500"
                  placeholder="Enter username or email"
                  required
                />
              </div>
            </div>

            <div>
              <label htmlFor="password" className="block text-sm font-medium text-gray-300 mb-2">
                Password
              </label>
              <div className="relative">
                <Lock className="absolute left-3 top-1/2 transform -translate-y-1/2 w-5 h-5 text-gray-400" />
                <input
                  id="password"
                  type={showPassword ? "text" : "password"}
                  name="password"
                  value={formData.password}
                  onChange={handleChange}
                  className="w-full pl-10 pr-12 py-2 bg-gray-700 border border-gray-600 rounded-lg text-white placeholder-gray-400 focus:outline-none focus:ring-2 focus:ring-blue-500"
                  placeholder="Enter password"
                  required
                />
                <button
                  type="button"
                  onClick={() => setShowPassword(!showPassword)}
                  className="absolute right-3 top-1/2 transform -translate-y-1/2 text-gray-400 hover:text-gray-300"
                >
                  {showPassword ? <EyeOff className="w-5 h-5" /> : <Eye className="w-5 h-5" />}
                </button>
              </div>
            </div>

            <div className="flex items-center justify-between">
              <Link
                to="/forgot-password"
                className="text-sm text-blue-400 hover:text-blue-300"
              >
                Forgot password?
              </Link>
            </div>

            {/* Cloudflare Turnstile */}
            {turnstileConfig.enabled && (
              <Turnstile
                siteKey={turnstileConfig.site_key}
                onVerify={setTurnstileToken}
                onError={() => setTurnstileToken('')}
                onExpire={() => setTurnstileToken('')}
                size="flexible"
                theme="auto"
              />
            )}

            <button
              type="submit"
              disabled={loading || (turnstileConfig.enabled && !turnstileToken)}
              className="w-full bg-blue-600 hover:bg-blue-700 text-white font-medium py-2 px-4 rounded-lg transition disabled:opacity-50 disabled:cursor-not-allowed"
            >
              {loading ? 'Signing in...' : 'Sign In'}
            </button>
          </form>

          {appSettings.signup_enabled && (
            <div className="mt-6 text-center">
              <p className="text-gray-400 text-sm">
                Don't have an account?{' '}
                <Link to="/register" className="text-blue-400 hover:text-blue-300 font-medium">
                  Sign up
                </Link>
              </p>
            </div>
          )}
        </div>

        {/* Footer */}
        <div className="mt-8 text-center text-gray-500 text-sm">
          {/* <p>&copy; 2025 {appSettings.app_name}. All rights reserved.</p> */}
          <p>&copy; 2025 Do Code Indonesia. All rights reserved.</p>
        </div>
      </div>
    </div>
  );
};

export default Login;
