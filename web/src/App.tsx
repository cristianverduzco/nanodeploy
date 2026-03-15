import React, { useEffect, useState } from 'react';
import axios from 'axios';
import { Database, Plus, Trash2, RefreshCw, Server, Activity } from 'lucide-react';

const API_BASE = 'http://localhost:9090/api/v1';

type ServicePhase = 'Pending' | 'Provisioning' | 'Ready' | 'Failed' | 'Terminating';
type ServiceType = 'postgresql' | 'redis' | 'rabbitmq';

interface ManagedService {
  metadata: {
    name: string;
    namespace: string;
    creationTimestamp: string;
  };
  spec: {
    type: ServiceType;
    version: string;
    replicas: number;
    storageGB: number;
    databaseName?: string;
  };
  status: {
    phase: ServicePhase;
    message: string;
    endpoint: string;
    lastUpdated: string;
  };
}

interface CreateServiceForm {
  name: string;
  namespace: string;
  type: ServiceType;
  version: string;
  replicas: number;
  storageGB: number;
  databaseName: string;
}

const defaultForm: CreateServiceForm = {
  name: '',
  namespace: 'default',
  type: 'postgresql',
  version: '15',
  replicas: 1,
  storageGB: 5,
  databaseName: '',
};

const phaseColors: Record<ServicePhase, string> = {
  Pending: 'bg-yellow-100 text-yellow-800',
  Provisioning: 'bg-blue-100 text-blue-800',
  Ready: 'bg-green-100 text-green-800',
  Failed: 'bg-red-100 text-red-800',
  Terminating: 'bg-gray-100 text-gray-800',
};

const serviceIcons: Record<ServiceType, string> = {
  postgresql: '🐘',
  redis: '🔴',
  rabbitmq: '🐇',
};

export default function App() {
  const [services, setServices] = useState<ManagedService[]>([]);
  const [loading, setLoading] = useState(false);
  const [showForm, setShowForm] = useState(false);
  const [form, setForm] = useState<CreateServiceForm>(defaultForm);
  const [error, setError] = useState<string | null>(null);

  const fetchServices = async () => {
    setLoading(true);
    setError(null);
    try {
      const res = await axios.get(`${API_BASE}/services`);
      setServices(res.data.items || []);
    } catch (e) {
      setError('Failed to connect to NanoDeploy API. Is the operator running?');
    } finally {
      setLoading(false);
    }
  };

  useEffect(() => {
    fetchServices();
    const interval = setInterval(fetchServices, 10000);
    return () => clearInterval(interval);
  }, []);

  const createService = async () => {
    try {
      await axios.post(`${API_BASE}/services`, form);
      setShowForm(false);
      setForm(defaultForm);
      fetchServices();
    } catch (e) {
      setError('Failed to create service.');
    }
  };

  const deleteService = async (namespace: string, name: string) => {
    if (!window.confirm(`Delete ${name}?`)) return;
    try {
      await axios.delete(`${API_BASE}/services/${namespace}/${name}`);
      fetchServices();
    } catch (e) {
      setError('Failed to delete service.');
    }
  };

  return (
    <div className="min-h-screen bg-gray-950 text-gray-100">
      {/* Header */}
      <header className="border-b border-gray-800 px-6 py-4">
        <div className="max-w-6xl mx-auto flex items-center justify-between">
          <div className="flex items-center gap-3">
            <Server className="text-indigo-400" size={28} />
            <div>
              <h1 className="text-xl font-bold text-white">NanoDeploy</h1>
              <p className="text-xs text-gray-400">Managed Service Infra Orchestration</p>
            </div>
          </div>
          <div className="flex items-center gap-3">
            <button
              onClick={fetchServices}
              className="p-2 rounded-lg hover:bg-gray-800 text-gray-400 hover:text-white transition"
            >
              <RefreshCw size={18} className={loading ? 'animate-spin' : ''} />
            </button>
            <button
              onClick={() => setShowForm(true)}
              className="flex items-center gap-2 bg-indigo-600 hover:bg-indigo-500 text-white px-4 py-2 rounded-lg text-sm font-medium transition"
            >
              <Plus size={16} />
              New Service
            </button>
          </div>
        </div>
      </header>

      <main className="max-w-6xl mx-auto px-6 py-8">
        {/* Error */}
        {error && (
          <div className="mb-6 bg-red-900/30 border border-red-700 text-red-300 px-4 py-3 rounded-lg text-sm">
            {error}
          </div>
        )}

        {/* Stats */}
        <div className="grid grid-cols-3 gap-4 mb-8">
          {[
            { label: 'Total Services', value: services.length, icon: <Database size={20} /> },
            { label: 'Ready', value: services.filter(s => s.status?.phase === 'Ready').length, icon: <Activity size={20} /> },
            { label: 'Namespaces', value: [...new Set(services.map(s => s.metadata.namespace))].length, icon: <Server size={20} /> },
          ].map((stat) => (
            <div key={stat.label} className="bg-gray-900 border border-gray-800 rounded-xl p-4 flex items-center gap-4">
              <div className="text-indigo-400">{stat.icon}</div>
              <div>
                <p className="text-2xl font-bold text-white">{stat.value}</p>
                <p className="text-xs text-gray-400">{stat.label}</p>
              </div>
            </div>
          ))}
        </div>

        {/* Services Table */}
        <div className="bg-gray-900 border border-gray-800 rounded-xl overflow-hidden">
          <div className="px-6 py-4 border-b border-gray-800">
            <h2 className="font-semibold text-white">Managed Services</h2>
          </div>
          {services.length === 0 ? (
            <div className="px-6 py-16 text-center text-gray-500">
              <Database size={40} className="mx-auto mb-3 opacity-30" />
              <p>No services yet. Deploy your first one!</p>
            </div>
          ) : (
            <table className="w-full">
              <thead>
                <tr className="text-xs text-gray-400 uppercase border-b border-gray-800">
                  <th className="px-6 py-3 text-left">Name</th>
                  <th className="px-6 py-3 text-left">Type</th>
                  <th className="px-6 py-3 text-left">Namespace</th>
                  <th className="px-6 py-3 text-left">Phase</th>
                  <th className="px-6 py-3 text-left">Endpoint</th>
                  <th className="px-6 py-3 text-left">Replicas</th>
                  <th className="px-6 py-3 text-left"></th>
                </tr>
              </thead>
              <tbody>
                {services.map((svc) => (
                  <tr key={`${svc.metadata.namespace}/${svc.metadata.name}`} className="border-b border-gray-800 hover:bg-gray-800/50 transition">
                    <td className="px-6 py-4 font-medium text-white">{svc.metadata.name}</td>
                    <td className="px-6 py-4">
                      <span className="flex items-center gap-2 text-sm">
                        {serviceIcons[svc.spec.type]} {svc.spec.type} {svc.spec.version}
                      </span>
                    </td>
                    <td className="px-6 py-4 text-sm text-gray-400">{svc.metadata.namespace}</td>
                    <td className="px-6 py-4">
                      <span className={`text-xs px-2 py-1 rounded-full font-medium ${phaseColors[svc.status?.phase] || 'bg-gray-700 text-gray-300'}`}>
                        {svc.status?.phase || 'Unknown'}
                      </span>
                    </td>
                    <td className="px-6 py-4 text-xs text-gray-400 font-mono">{svc.status?.endpoint || '—'}</td>
                    <td className="px-6 py-4 text-sm text-gray-400">{svc.spec.replicas}</td>
                    <td className="px-6 py-4">
                      <button
                        onClick={() => deleteService(svc.metadata.namespace, svc.metadata.name)}
                        className="text-gray-500 hover:text-red-400 transition"
                      >
                        <Trash2 size={16} />
                      </button>
                    </td>
                  </tr>
                ))}
              </tbody>
            </table>
          )}
        </div>
      </main>

      {/* Create Service Modal */}
      {showForm && (
        <div className="fixed inset-0 bg-black/60 flex items-center justify-center z-50">
          <div className="bg-gray-900 border border-gray-700 rounded-2xl p-6 w-full max-w-md">
            <h2 className="text-lg font-semibold text-white mb-6">Deploy New Service</h2>
            <div className="space-y-4">
              {[
                { label: 'Name', key: 'name', type: 'text', placeholder: 'my-postgres' },
                { label: 'Namespace', key: 'namespace', type: 'text', placeholder: 'default' },
                { label: 'Version', key: 'version', type: 'text', placeholder: '15' },
                { label: 'Database Name', key: 'databaseName', type: 'text', placeholder: 'appdb' },
              ].map(({ label, key, type, placeholder }) => (
                <div key={key}>
                  <label className="block text-sm text-gray-400 mb-1">{label}</label>
                  <input
                    type={type}
                    placeholder={placeholder}
                    value={form[key as keyof CreateServiceForm] as string}
                    onChange={e => setForm({ ...form, [key]: e.target.value })}
                    className="w-full bg-gray-800 border border-gray-700 rounded-lg px-3 py-2 text-sm text-white focus:outline-none focus:border-indigo-500"
                  />
                </div>
              ))}
              <div>
                <label className="block text-sm text-gray-400 mb-1">Type</label>
                <select
                  value={form.type}
                  onChange={e => setForm({ ...form, type: e.target.value as ServiceType })}
                  className="w-full bg-gray-800 border border-gray-700 rounded-lg px-3 py-2 text-sm text-white focus:outline-none focus:border-indigo-500"
                >
                  <option value="postgresql">PostgreSQL</option>
                  <option value="redis">Redis</option>
                  <option value="rabbitmq">RabbitMQ</option>
                </select>
              </div>
              <div className="grid grid-cols-2 gap-3">
                <div>
                  <label className="block text-sm text-gray-400 mb-1">Replicas</label>
                  <input
                    type="number"
                    min={1}
                    value={form.replicas}
                    onChange={e => setForm({ ...form, replicas: parseInt(e.target.value) })}
                    className="w-full bg-gray-800 border border-gray-700 rounded-lg px-3 py-2 text-sm text-white focus:outline-none focus:border-indigo-500"
                  />
                </div>
                <div>
                  <label className="block text-sm text-gray-400 mb-1">Storage (GB)</label>
                  <input
                    type="number"
                    min={1}
                    value={form.storageGB}
                    onChange={e => setForm({ ...form, storageGB: parseInt(e.target.value) })}
                    className="w-full bg-gray-800 border border-gray-700 rounded-lg px-3 py-2 text-sm text-white focus:outline-none focus:border-indigo-500"
                  />
                </div>
              </div>
            </div>
            <div className="flex gap-3 mt-6">
              <button
                onClick={() => setShowForm(false)}
                className="flex-1 bg-gray-800 hover:bg-gray-700 text-white py-2 rounded-lg text-sm transition"
              >
                Cancel
              </button>
              <button
                onClick={createService}
                className="flex-1 bg-indigo-600 hover:bg-indigo-500 text-white py-2 rounded-lg text-sm font-medium transition"
              >
                Deploy
              </button>
            </div>
          </div>
        </div>
      )}
    </div>
  );
}