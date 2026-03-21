import { useState } from 'react'
import type { FormEvent } from 'react'
import { useMutation, useQuery } from '@tanstack/react-query'
import api from '../lib/api'
import type {
  DownloadClientConfig,
  GeneralConfig,
  IndexerConfig,
  NotificationConfig,
  QualityDefinition,
  QualityProfile,
  SchemaOption,
} from '../lib/types'

const tabs = ['general', 'logging', 'quality', 'indexers', 'clients', 'notifications', 'security'] as const

type Tab = (typeof tabs)[number]

function Section({ title, subtitle, children }: { title: string; subtitle?: string; children: React.ReactNode }) {
  return (
    <section className="panel p-5">
      <div className="mb-4">
        <h3 className="text-lg font-semibold text-gray-100">{title}</h3>
        {subtitle ? <p className="mt-1 text-sm text-gray-500">{subtitle}</p> : null}
      </div>
      {children}
    </section>
  )
}

export default function Settings() {
  const [activeTab, setActiveTab] = useState<Tab>('general')
  const [newApiKey, setNewApiKey] = useState<string | null>(null)
  const [apiKeyError, setApiKeyError] = useState<string | null>(null)
  const [logTarget, setLogTarget] = useState<string>('')
  const [testLogResult, setTestLogResult] = useState<string | null>(null)

  const hostConfig = useQuery<GeneralConfig>({ queryKey: ['host-config'], queryFn: () => api.get('/config/host').then((res) => res.data) })
  const qualityProfiles = useQuery<QualityProfile[]>({ queryKey: ['quality-profiles'], queryFn: () => api.get('/qualityprofile').then((res) => res.data) })
  const qualityDefinitions = useQuery<QualityDefinition[]>({ queryKey: ['quality-definitions'], queryFn: () => api.get('/qualitydefinition').then((res) => res.data) })
  const indexers = useQuery<IndexerConfig[]>({ queryKey: ['indexers'], queryFn: () => api.get('/indexer').then((res) => res.data) })
  const indexerSchemas = useQuery<SchemaOption[]>({ queryKey: ['indexer-schemas'], queryFn: () => api.get('/indexer/schema').then((res) => res.data) })
  const downloadClients = useQuery<DownloadClientConfig[]>({ queryKey: ['download-clients'], queryFn: () => api.get('/downloadclient').then((res) => res.data) })
  const clientSchemas = useQuery<SchemaOption[]>({ queryKey: ['client-schemas'], queryFn: () => api.get('/downloadclient/schema').then((res) => res.data) })
  const notifications = useQuery<NotificationConfig[]>({ queryKey: ['notifications'], queryFn: () => api.get('/notification').then((res) => res.data) })
  const notificationSchemas = useQuery<SchemaOption[]>({ queryKey: ['notification-schemas'], queryFn: () => api.get('/notification/schema').then((res) => res.data) })

  const saveGeneral = useMutation({
    mutationFn: (payload: Partial<GeneralConfig>) => api.put('/config/host', payload),
  })

  const saveLogging = useMutation({
    mutationFn: (payload: Partial<GeneralConfig>) => api.put('/config/host', payload),
    onSuccess: () => setTestLogResult(null),
  })

  const testLogs = useMutation({
    mutationFn: () => api.post<{ message: string; target: string }>('/system/logs/test').then((r) => r.data),
    onSuccess: (data) => setTestLogResult(`✅ ${data.message} → target: ${data.target}`),
    onError: (err: Error) => setTestLogResult(`❌ ${err.message}`),
  })

  const regenerateKey = useMutation({
    mutationFn: () => api.post<{ apiKey: string }>('/auth/apikey/regenerate').then((res) => res.data),
    onSuccess: (data) => {
      setNewApiKey(data.apiKey)
      setApiKeyError(null)
    },
    onError: (err: Error) => {
      setApiKeyError(err.message)
    },
  })

  function submitGeneral(event: FormEvent<HTMLFormElement>) {
    event.preventDefault()
    const form = new FormData(event.currentTarget)
    saveGeneral.mutate({
      bindAddress: String(form.get('bindAddress') ?? ''),
      port: Number(form.get('port') ?? 7878),
      urlBase: String(form.get('urlBase') ?? ''),
      logLevel: String(form.get('logLevel') ?? 'info'),
    })
  }

  function submitLogging(event: FormEvent<HTMLFormElement>) {
    event.preventDefault()
    const form = new FormData(event.currentTarget)
    saveLogging.mutate({
      logLevel: String(form.get('logLevel') ?? 'info'),
      logTarget: String(form.get('logTarget') ?? 'stderr'),
      logFile: String(form.get('logFile') ?? ''),
      syslogAddress: String(form.get('syslogAddress') ?? ''),
      syslogPort: Number(form.get('syslogPort') ?? 514),
      syslogNetwork: String(form.get('syslogNetwork') ?? 'udp'),
    })
  }

  const effectiveLogTarget = logTarget || hostConfig.data?.logTarget || 'stderr'

  return (
    <div className="grid gap-6 lg:grid-cols-[220px_minmax(0,1fr)]">
      <nav className="panel flex flex-row gap-2 overflow-x-auto p-3 lg:flex-col lg:overflow-visible">
        {tabs.map((tab) => (
          <button key={tab} onClick={() => setActiveTab(tab)} className={activeTab === tab ? 'btn-primary justify-start' : 'btn-secondary justify-start'}>
            {tab[0].toUpperCase() + tab.slice(1)}
          </button>
        ))}
      </nav>

      <div className="space-y-6">
        {activeTab === 'general' ? (
          <Section title="Host settings" subtitle="Update bind address, port, and operational defaults.">
            <form className="space-y-4" onSubmit={submitGeneral}>
              <div className="grid gap-4 md:grid-cols-2">
                <label className="text-sm text-gray-400">Bind address<input name="bindAddress" className="field mt-2" defaultValue={hostConfig.data?.bindAddress ?? ''} /></label>
                <label className="text-sm text-gray-400">Port<input name="port" className="field mt-2" type="number" defaultValue={hostConfig.data?.port ?? 7878} /></label>
                <label className="text-sm text-gray-400">URL base<input name="urlBase" className="field mt-2" defaultValue={hostConfig.data?.urlBase ?? ''} /></label>
                <label className="text-sm text-gray-400">Log level<input name="logLevel" className="field mt-2" defaultValue={hostConfig.data?.logLevel ?? 'info'} /></label>
              </div>
              <div className="flex justify-end">
                <button type="submit" className="btn-primary" disabled={saveGeneral.isPending}>Save host config</button>
              </div>
            </form>
          </Section>
        ) : null}

        {activeTab === 'logging' ? (
          <Section title="Logging" subtitle="Configure where GoRadarr sends structured JSON log output.">
            <form className="space-y-5" onSubmit={submitLogging}>
              <div className="grid gap-4 md:grid-cols-2">
                <label className="text-sm text-gray-400">
                  Log level
                  <select name="logLevel" className="field mt-2" defaultValue={hostConfig.data?.logLevel ?? 'info'}>
                    <option value="debug">debug</option>
                    <option value="info">info</option>
                    <option value="warn">warn</option>
                    <option value="error">error</option>
                  </select>
                </label>
                <label className="text-sm text-gray-400">
                  Log target
                  <select
                    name="logTarget"
                    className="field mt-2"
                    value={effectiveLogTarget}
                    onChange={(e) => setLogTarget(e.target.value)}
                  >
                    <option value="stderr">stderr (default)</option>
                    <option value="stdout">stdout</option>
                    <option value="file">file</option>
                    <option value="syslog">syslog (remote)</option>
                  </select>
                </label>
              </div>

              {effectiveLogTarget === 'file' ? (
                <label className="block text-sm text-gray-400">
                  Log file path
                  <input name="logFile" className="field mt-2" defaultValue={hostConfig.data?.logFile ?? ''} placeholder="/config/goradarr.log" />
                </label>
              ) : null}

              {effectiveLogTarget === 'syslog' ? (
                <div className="rounded-lg border border-gray-700 bg-gray-800/50 p-4 space-y-4">
                  <p className="text-xs font-semibold uppercase tracking-wider text-gray-500">Remote syslog (RFC 3164)</p>
                  <div className="grid gap-4 md:grid-cols-3">
                    <label className="text-sm text-gray-400">
                      Host / IP
                      <input
                        name="syslogAddress"
                        className="field mt-2 font-mono"
                        defaultValue={hostConfig.data?.syslogAddress ?? ''}
                        placeholder="192.168.10.152"
                      />
                    </label>
                    <label className="text-sm text-gray-400">
                      Port
                      <input
                        name="syslogPort"
                        className="field mt-2"
                        type="number"
                        defaultValue={hostConfig.data?.syslogPort ?? 514}
                        placeholder="514"
                      />
                    </label>
                    <label className="text-sm text-gray-400">
                      Protocol
                      <select name="syslogNetwork" className="field mt-2" defaultValue={hostConfig.data?.syslogNetwork ?? 'udp'}>
                        <option value="udp">UDP</option>
                        <option value="tcp">TCP</option>
                        <option value="unix">Unix socket</option>
                      </select>
                    </label>
                  </div>
                  <p className="text-xs text-gray-500">
                    Each log entry is sent as a single structured JSON line inside the syslog message body.
                  </p>
                </div>
              ) : null}

              <div className="flex items-center justify-between gap-3">
                <div className="flex gap-3">
                  <button
                    type="button"
                    className="btn-secondary"
                    disabled={testLogs.isPending}
                    onClick={() => testLogs.mutate()}
                  >
                    {testLogs.isPending ? 'Sending…' : 'Send test logs'}
                  </button>
                  {testLogResult ? <span className="self-center text-sm text-gray-400">{testLogResult}</span> : null}
                </div>
                <button type="submit" className="btn-primary" disabled={saveLogging.isPending}>
                  {saveLogging.isPending ? 'Saving…' : 'Save logging config'}
                </button>
              </div>
            </form>
          </Section>
        ) : null}

        {activeTab === 'quality' ? (
          <>
            <Section title="Quality profiles" subtitle="Profiles available to movie records and automatic upgrades.">
              <div className="space-y-3">
                {qualityProfiles.data?.map((profile) => <div key={profile.id} className="panel-muted flex items-center justify-between px-4 py-3 text-sm"><span>{profile.name}</span><span className="text-gray-500">Cutoff {profile.cutoff}</span></div>)}
              </div>
            </Section>
            <Section title="Quality definitions">
              <div className="grid gap-3 md:grid-cols-2">
                {qualityDefinitions.data?.map((definition) => <div key={definition.id} className="panel-muted p-4 text-sm text-gray-300"><p className="font-medium text-gray-100">{definition.title}</p><p className="mt-2 text-gray-500">Preferred size {definition.preferredSize}</p></div>)}
              </div>
            </Section>
          </>
        ) : null}

        {activeTab === 'indexers' ? (
          <Section title="Indexers" subtitle="Configured sources and available templates.">
            <div className="grid gap-4 xl:grid-cols-2">
              <div className="space-y-3">{indexers.data?.map((item) => <div key={item.id} className="panel-muted p-4"><p className="font-medium text-gray-100">{item.name}</p><p className="mt-1 text-sm text-gray-500">{item.implementation} · priority {item.priority}</p></div>)}</div>
              <div className="space-y-3">{indexerSchemas.data?.map((item) => <div key={item.implementation} className="panel-muted p-4"><p className="font-medium text-gray-100">{item.implementation}</p><p className="mt-1 text-sm text-gray-500">{item.configContract}</p></div>)}</div>
            </div>
          </Section>
        ) : null}

        {activeTab === 'clients' ? (
          <Section title="Download clients" subtitle="Live clients on the left, supported templates on the right.">
            <div className="grid gap-4 xl:grid-cols-2">
              <div className="space-y-3">{downloadClients.data?.map((item) => <div key={item.id} className="panel-muted p-4"><p className="font-medium text-gray-100">{item.name}</p><p className="mt-1 text-sm text-gray-500">{item.implementation} · {item.protocol}</p></div>)}</div>
              <div className="space-y-3">{clientSchemas.data?.map((item) => <div key={item.implementation} className="panel-muted p-4"><p className="font-medium text-gray-100">{item.implementation}</p><p className="mt-1 text-sm text-gray-500">{item.protocol ?? 'mixed'} transport</p></div>)}</div>
            </div>
          </Section>
        ) : null}

        {activeTab === 'notifications' ? (
          <Section title="Notifications" subtitle="Delivery rules and available providers.">
            <div className="grid gap-4 xl:grid-cols-2">
              <div className="space-y-3">{notifications.data?.map((item) => <div key={item.id} className="panel-muted p-4"><p className="font-medium text-gray-100">{item.name}</p><p className="mt-1 text-sm text-gray-500">{item.implementation}</p></div>)}</div>
              <div className="space-y-3">{notificationSchemas.data?.map((item) => <div key={item.implementation} className="panel-muted p-4"><p className="font-medium text-gray-100">{item.implementation}</p><p className="mt-1 text-sm text-gray-500">{item.configContract}</p></div>)}</div>
            </div>
          </Section>
        ) : null}

        {activeTab === 'security' ? (
          <Section title="Security" subtitle="Manage API access credentials.">
            <div className="space-y-4">
              <div>
                <p className="text-sm text-gray-400">Current API Key</p>
                <p className="mt-2 font-mono text-sm text-gray-300">
                  ••••••••••••••••••••••••••••••••
                </p>
              </div>
              {newApiKey ? (
                <div>
                  <p className="mb-2 text-sm text-gray-400">New API Key (copy it now)</p>
                  <div className="flex gap-2">
                    <input readOnly className="field flex-1 font-mono text-sm" value={newApiKey} />
                    <button
                      className="btn-secondary"
                      onClick={() => navigator.clipboard.writeText(newApiKey)}
                    >
                      Copy
                    </button>
                  </div>
                </div>
              ) : null}
              {apiKeyError ? (
                <p className="text-sm text-red-400">{apiKeyError}</p>
              ) : null}
              <button
                className="btn-primary"
                disabled={regenerateKey.isPending}
                onClick={() => regenerateKey.mutate()}
              >
                {regenerateKey.isPending ? 'Regenerating…' : 'Regenerate API Key'}
              </button>
            </div>
          </Section>
        ) : null}
      </div>
    </div>
  )
}
