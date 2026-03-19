import { useState } from 'react';
import { clsx } from 'clsx';
import { Settings2, Layers, Rss, HardDrive, Bell } from 'lucide-react';

type Tab = 'general' | 'quality' | 'indexers' | 'clients' | 'notifications';

const tabs: { id: Tab; label: string; icon: React.ElementType }[] = [
  { id: 'general', label: 'General', icon: Settings2 },
  { id: 'quality', label: 'Quality', icon: Layers },
  { id: 'indexers', label: 'Indexers', icon: Rss },
  { id: 'clients', label: 'Download Clients', icon: HardDrive },
  { id: 'notifications', label: 'Notifications', icon: Bell },
];

function SectionCard({ title, children }: { title: string; children: React.ReactNode }) {
  return (
    <div className="bg-gray-900 border border-gray-800 rounded-xl p-5">
      <h3 className="text-sm font-semibold text-gray-300 mb-4">{title}</h3>
      {children}
    </div>
  );
}

function FieldRow({ label, children }: { label: string; children: React.ReactNode }) {
  return (
    <div className="flex items-center justify-between py-2.5 border-b border-gray-800/60 last:border-0">
      <label className="text-sm text-gray-400 w-48 shrink-0">{label}</label>
      <div className="flex-1">{children}</div>
    </div>
  );
}

function Toggle({ defaultChecked = false }: { defaultChecked?: boolean }) {
  const [on, setOn] = useState(defaultChecked);
  return (
    <button
      onClick={() => setOn(!on)}
      className={clsx(
        'relative inline-flex h-5 w-9 rounded-full transition-colors',
        on ? 'bg-yellow-400' : 'bg-gray-700'
      )}
    >
      <span className={clsx(
        'absolute top-0.5 left-0.5 w-4 h-4 rounded-full bg-white shadow transition-transform',
        on && 'translate-x-4'
      )} />
    </button>
  );
}

function TextInput({ placeholder = '', defaultValue = '' }: { placeholder?: string; defaultValue?: string }) {
  return (
    <input
      type="text"
      placeholder={placeholder}
      defaultValue={defaultValue}
      className="w-full bg-gray-800 border border-gray-700 rounded-lg px-3 py-1.5 text-sm text-gray-100 placeholder-gray-600 focus:outline-none focus:border-yellow-400/50 transition-colors"
    />
  );
}

function SelectInput({ options }: { options: string[] }) {
  return (
    <select className="bg-gray-800 border border-gray-700 rounded-lg px-3 py-1.5 text-sm text-gray-100 focus:outline-none focus:border-yellow-400/50 transition-colors">
      {options.map((o) => <option key={o}>{o}</option>)}
    </select>
  );
}

function GeneralTab() {
  return (
    <div className="space-y-4">
      <SectionCard title="Host">
        <FieldRow label="Bind Address"><TextInput defaultValue="*" /></FieldRow>
        <FieldRow label="Port"><TextInput defaultValue="7878" /></FieldRow>
        <FieldRow label="URL Base"><TextInput placeholder="/" /></FieldRow>
        <FieldRow label="Enable SSL"><Toggle /></FieldRow>
      </SectionCard>
      <SectionCard title="Security">
        <FieldRow label="Authentication Method">
          <SelectInput options={['None', 'Basic (Browser Popup)', 'Forms (Login Page)']} />
        </FieldRow>
      </SectionCard>
      <SectionCard title="Logging">
        <FieldRow label="Log Level">
          <SelectInput options={['Trace', 'Debug', 'Info', 'Warn', 'Error']} />
        </FieldRow>
      </SectionCard>
      <SectionCard title="Analytics">
        <FieldRow label="Send Anonymous Usage Data"><Toggle defaultChecked /></FieldRow>
      </SectionCard>
    </div>
  );
}

function QualityTab() {
  return (
    <div className="space-y-4">
      <SectionCard title="Quality Profiles">
        <p className="text-sm text-gray-500 mb-3">Define acceptable quality levels for movies.</p>
        <div className="space-y-2">
          {['Any', 'SD', '720p', '1080p', '2160p (4K)'].map((p) => (
            <div key={p} className="flex items-center justify-between px-3 py-2 bg-gray-800 rounded-lg">
              <span className="text-sm text-gray-200">{p}</span>
              <div className="flex gap-2">
                <button className="text-xs text-gray-500 hover:text-gray-300 transition-colors">Edit</button>
                <button className="text-xs text-red-500 hover:text-red-400 transition-colors">Delete</button>
              </div>
            </div>
          ))}
        </div>
        <button className="mt-3 text-sm text-yellow-400 hover:text-yellow-300 transition-colors">
          + Add Profile
        </button>
      </SectionCard>
    </div>
  );
}

function IndexersTab() {
  return (
    <div className="space-y-4">
      <SectionCard title="Indexers">
        <p className="text-sm text-gray-500 mb-3">Configure Usenet and torrent indexers.</p>
        <div className="text-center py-8 text-gray-600">
          <Rss size={32} className="mx-auto mb-2 opacity-30" />
          <p className="text-sm">No indexers configured</p>
          <button className="mt-3 text-sm text-yellow-400 hover:text-yellow-300 transition-colors">
            + Add Indexer
          </button>
        </div>
      </SectionCard>
      <SectionCard title="Options">
        <FieldRow label="Minimum Age"><TextInput defaultValue="0" /></FieldRow>
        <FieldRow label="Retention"><TextInput defaultValue="0" /></FieldRow>
        <FieldRow label="RSS Sync Interval"><TextInput defaultValue="60" /></FieldRow>
      </SectionCard>
    </div>
  );
}

function DownloadClientsTab() {
  return (
    <div className="space-y-4">
      <SectionCard title="Download Clients">
        <p className="text-sm text-gray-500 mb-3">Configure download clients for Usenet and torrents.</p>
        <div className="text-center py-8 text-gray-600">
          <HardDrive size={32} className="mx-auto mb-2 opacity-30" />
          <p className="text-sm">No download clients configured</p>
          <button className="mt-3 text-sm text-yellow-400 hover:text-yellow-300 transition-colors">
            + Add Client
          </button>
        </div>
      </SectionCard>
      <SectionCard title="Completed Download Handling">
        <FieldRow label="Enable"><Toggle defaultChecked /></FieldRow>
        <FieldRow label="Remove Completed"><Toggle /></FieldRow>
      </SectionCard>
    </div>
  );
}

function NotificationsTab() {
  return (
    <div className="space-y-4">
      <SectionCard title="Connections">
        <p className="text-sm text-gray-500 mb-3">Configure notifications for grab, download, and rename events.</p>
        <div className="text-center py-8 text-gray-600">
          <Bell size={32} className="mx-auto mb-2 opacity-30" />
          <p className="text-sm">No connections configured</p>
          <button className="mt-3 text-sm text-yellow-400 hover:text-yellow-300 transition-colors">
            + Add Connection
          </button>
        </div>
      </SectionCard>
    </div>
  );
}

export default function Settings() {
  const [activeTab, setActiveTab] = useState<Tab>('general');

  const tabContent: Record<Tab, React.ReactNode> = {
    general: <GeneralTab />,
    quality: <QualityTab />,
    indexers: <IndexersTab />,
    clients: <DownloadClientsTab />,
    notifications: <NotificationsTab />,
  };

  return (
    <div className="p-4 md:p-6">
      <div className="mb-6">
        <h1 className="text-2xl font-bold text-gray-100">Settings</h1>
      </div>

      <div className="flex gap-6 flex-col lg:flex-row">
        {/* Tabs sidebar */}
        <nav className="flex lg:flex-col gap-1 overflow-x-auto lg:overflow-visible shrink-0 lg:w-44">
          {tabs.map(({ id, label, icon: Icon }) => (
            <button
              key={id}
              onClick={() => setActiveTab(id)}
              className={clsx(
                'flex items-center gap-2.5 px-3 py-2.5 rounded-lg text-sm font-medium whitespace-nowrap transition-colors text-left',
                activeTab === id
                  ? 'bg-yellow-400/10 text-yellow-400'
                  : 'text-gray-400 hover:bg-gray-800 hover:text-gray-200'
              )}
            >
              <Icon size={16} className="shrink-0" />
              {label}
            </button>
          ))}
        </nav>

        {/* Content */}
        <div className="flex-1 min-w-0">
          {tabContent[activeTab]}
          <div className="mt-4 flex justify-end">
            <button className="bg-yellow-400 hover:bg-yellow-300 text-gray-900 font-medium px-6 py-2 rounded-lg text-sm transition-colors">
              Save Changes
            </button>
          </div>
        </div>
      </div>
    </div>
  );
}
