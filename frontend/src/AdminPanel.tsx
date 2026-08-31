import { useEffect, useState } from 'react'
import { Bar, BarChart, CartesianGrid, Cell, Legend, Pie, PieChart, ResponsiveContainer, Tooltip, XAxis, YAxis } from 'recharts'
import { AdminFile, AdminStats, api } from './api'

const bytes = (n: number) => n < 1024 ? `${n} B` : n < 1048576 ? `${(n / 1024).toFixed(1)} KB` : `${(n / 1048576).toFixed(1)} MB`
const COLORS = ['#147f83', '#8bc7bd']

export default function AdminPanel({ visible, onBack }: { visible: boolean; onBack?: () => void }) {
  const [stats, setStats] = useState<AdminStats | null>(null), [files, setFiles] = useState<AdminFile[]>([]), [next, setNext] = useState(''), [email, setEmail] = useState(''), [message, setMessage] = useState(''), [busy, setBusy] = useState(false), [adminSection, setAdminSection] = useState<'home' | 'files' | 'stats'>('home'), [shareFileId, setShareFileId] = useState('')
  const refresh = async (after = '') => { const [s, f] = await Promise.all([api.adminStats(), api.adminFiles(after)]); setStats(s); setFiles(f.items); setNext(f.hasNextPage ? f.nextCursor ?? '' : '') }
  useEffect(() => {
    if (!visible || !onBack) return
    void refresh().catch(() => setMessage('Unable to load administrator data'))
    const events = new EventSource('/api/events/downloads')
    events.onmessage = () => void refresh().catch(() => setMessage('Live statistics temporarily unavailable'))
    return () => events.close()
  }, [visible])
  useEffect(() => {
    if (!visible || !onBack) return
    const panel = document.querySelector<HTMLElement>('.admin-panel')
    if (!panel) return
    const eyebrow = panel.querySelector<HTMLElement>('.panel-head .eyebrow')
    const heading = panel.querySelector<HTMLElement>('.panel-head h2')
    if (eyebrow) eyebrow.textContent = 'STATISTICS & OVERSIGHT'
    if (heading) heading.textContent = 'Workspace analytics'
    if (onBack && !panel.querySelector('.admin-back')) {
      const button = document.createElement('button'); button.className = 'secondary admin-back'; button.textContent = 'Back to vault'; button.onclick = onBack
      panel.querySelector('.panel-head')?.prepend(button)
    }
  }, [visible, onBack])
  useEffect(() => {
    if (!visible) return
    const panel = document.querySelector<HTMLElement>('.admin-panel')
    if (!panel) return
    let tabs = panel.querySelector<HTMLElement>('.admin-tabs')
    if (!tabs) { tabs = document.createElement('div'); tabs.className = 'admin-tabs'; panel.querySelector('.panel-head')?.insertAdjacentElement('afterend', tabs) }
    tabs.innerHTML = ''
    const homeTab = document.createElement('button'); homeTab.textContent = 'Admin home'; homeTab.className = adminSection === 'home' ? 'active' : ''; homeTab.onclick = () => setAdminSection('home')
    const filesTab = document.createElement('button'); filesTab.textContent = 'All files'; filesTab.className = adminSection === 'files' ? 'active' : ''; filesTab.onclick = () => setAdminSection('files')
    const statsTab = document.createElement('button'); statsTab.textContent = 'Statistics'; statsTab.className = adminSection === 'stats' ? 'active' : ''; statsTab.onclick = () => setAdminSection('stats')
    tabs.append(homeTab, filesTab, statsTab)
    panel.querySelector('.stats')?.classList.add('admin-metrics'); panel.querySelector('.admin-charts')?.classList.add('admin-metrics'); panel.querySelector('.share-result')?.classList.add('admin-share'); panel.querySelector('.file-list')?.classList.add('admin-files')
    panel.querySelectorAll<HTMLElement>('.admin-metrics').forEach(element => { element.style.display = adminSection === 'stats' ? '' : 'none' })
    panel.querySelectorAll<HTMLElement>('.admin-files').forEach(element => { element.style.display = adminSection === 'files' ? '' : 'none' })
    panel.querySelectorAll<HTMLElement>('.admin-share').forEach(element => { element.style.display = adminSection === 'home' ? '' : 'none' })
    const upload = panel.querySelector<HTMLElement>('.panel-head .upload'); if (upload) upload.style.display = adminSection === 'home' ? '' : 'none'
    const heading = panel.querySelector<HTMLElement>('.panel-head h2'); if (heading) heading.textContent = adminSection === 'home' ? 'Admin home' : adminSection === 'files' ? 'All uploaded files' : 'Workspace analytics'
    const shareBox = panel.querySelector<HTMLElement>('.admin-share')
    if (shareBox && !shareBox.querySelector('.admin-file-picker')) {
      const select = document.createElement('select'); select.className = 'admin-file-picker'; select.innerHTML = '<option value="">Choose a file to share</option>' + files.map(file => `<option value="${file.id}">${file.name}</option>`).join(''); select.onchange = () => setShareFileId(select.value)
      const button = document.createElement('button'); button.className = 'primary admin-share-button'; button.textContent = 'Share selected file'; button.onclick = () => { if (shareFileId) void share(shareFileId) }
      shareBox.append(select, button)
    }
    const picker = shareBox?.querySelector<HTMLSelectElement>('.admin-file-picker'); if (picker) { picker.innerHTML = '<option value="">Choose a file to share</option>' + files.map(file => `<option value="${file.id}">${file.name}</option>`).join(''); picker.value = shareFileId }
    const shareButton = shareBox?.querySelector<HTMLButtonElement>('.admin-share-button'); if (shareButton) shareButton.disabled = !shareFileId || busy
  }, [visible, adminSection, stats, files, shareFileId, busy])
  const upload = async (file: File) => { setBusy(true); try { const csrf = await api.csrf(); const result = await api.upload(file, csrf.csrfToken); setMessage(result.results[0]?.status === 'created' ? 'File uploaded.' : 'Upload rejected.'); await refresh() } catch { setMessage('Administrator upload failed') } finally { setBusy(false) } }
  const share = async (fileId: string) => { if (!email.trim()) { setMessage('Enter a recipient email first'); return } setBusy(true); try { const csrf = await api.csrf(); await api.directShare(fileId, email.trim(), 'download', csrf.csrfToken); setMessage(`File shared with ${email.trim()}`); setEmail('') } catch { setMessage('Unable to share file with that user') } finally { setBusy(false) } }
  if (!visible || !onBack) return null
  const savings = stats ? Math.max(0, stats.logicalBytes - stats.physicalBytes) : 0
  const storageData = stats ? [{ name: 'Logical', bytes: stats.logicalBytes }, { name: 'Physical', bytes: stats.physicalBytes }] : []
  const compositionData = stats ? [{ name: 'Physical storage', value: stats.physicalBytes }, { name: 'Deduplicated savings', value: savings }] : []
  const activityData = stats ? [{ name: 'Users', value: stats.users }, { name: 'Files', value: stats.files }, { name: 'Downloads', value: stats.downloads }] : []
  return <section className="file-panel admin-panel"><div className="panel-head"><div><p className="eyebrow">ADMINISTRATION</p><h2>Vault oversight</h2><span className="muted">Live workspace health and file activity</span></div><label className="upload">{busy ? 'Working…' : 'Upload file'}<input type="file" disabled={busy} onChange={e => { const file = e.target.files?.[0]; if (file) void upload(file) }} /></label></div>{message && <p className="muted">{message}</p>}<div className="stats"><div className="stat"><span>Users</span><strong>{stats?.users ?? '—'}</strong><small>active accounts</small></div><div className="stat"><span>Files</span><strong>{stats?.files ?? '—'}</strong><small>workspace files</small></div><div className="stat"><span>Logical storage</span><strong>{stats ? bytes(stats.logicalBytes) : '—'}</strong><small>before deduplication</small></div><div className="stat"><span>Physical storage</span><strong>{stats ? bytes(stats.physicalBytes) : '—'}</strong><small>after deduplication</small></div><div className="stat"><span>Deduplication savings</span><strong>{stats ? bytes(savings) : '—'}</strong><small>{stats && stats.logicalBytes ? `${((savings / stats.logicalBytes) * 100).toFixed(1)}% saved` : '—'}</small></div><div className="stat"><span>Downloads</span><strong>{stats?.downloads ?? '—'}</strong><small>recorded events</small></div></div><div className="admin-charts"><article className="chart-card"><div><p className="eyebrow">STORAGE EFFICIENCY</p><h3>Logical vs physical usage</h3></div><ResponsiveContainer width="100%" height={220}><BarChart data={storageData} margin={{ top: 12, right: 12, left: 0, bottom: 4 }}><CartesianGrid strokeDasharray="3 3" stroke="#e4edf0" vertical={false} /><XAxis dataKey="name" tick={{ fill: '#71849a', fontSize: 11 }} /><YAxis tick={{ fill: '#71849a', fontSize: 11 }} tickFormatter={value => bytes(Number(value))} /><Tooltip formatter={value => bytes(Number(value))} contentStyle={{ borderRadius: 8, border: '1px solid #dce7ee' }} /><Bar dataKey="bytes" name="Storage" fill="#147f83" radius={[6, 6, 0, 0]} /></BarChart></ResponsiveContainer></article><article className="chart-card"><div><p className="eyebrow">DEDUPLICATION</p><h3>Physical storage footprint</h3></div><ResponsiveContainer width="100%" height={220}><PieChart><Pie data={compositionData} dataKey="value" nameKey="name" innerRadius={58} outerRadius={82} paddingAngle={4}>{compositionData.map((entry, index) => <Cell key={entry.name} fill={COLORS[index % COLORS.length]} />)}</Pie><Tooltip formatter={value => bytes(Number(value))} contentStyle={{ borderRadius: 8, border: '1px solid #dce7ee' }} /><Legend iconType="circle" wrapperStyle={{ fontSize: 11, color: '#71849a' }} /></PieChart></ResponsiveContainer></article><article className="chart-card chart-card-wide"><div><p className="eyebrow">WORKSPACE ACTIVITY</p><h3>Current platform totals</h3></div><ResponsiveContainer width="100%" height={220}><BarChart data={activityData} margin={{ top: 12, right: 12, left: 0, bottom: 4 }}><CartesianGrid strokeDasharray="3 3" stroke="#e4edf0" vertical={false} /><XAxis dataKey="name" tick={{ fill: '#71849a', fontSize: 11 }} /><YAxis allowDecimals={false} tick={{ fill: '#71849a', fontSize: 11 }} /><Tooltip contentStyle={{ borderRadius: 8, border: '1px solid #dce7ee' }} /><Bar dataKey="value" name="Count" fill="#72b8ab" radius={[6, 6, 0, 0]} /></BarChart></ResponsiveContainer></article></div><div className="share-result"><small>Share a file with a user</small><input value={email} onChange={e => setEmail(e.target.value)} placeholder="recipient@example.com" type="email" /></div>{files.length === 0 ? <div className="empty"><h3>No files yet</h3><p>Uploaded files will appear in the administrator inventory.</p></div> : <div className="file-list">{files.map(file => <div className="file-row" key={file.id}><div className="file-name"><strong>{file.name}</strong><small>{file.uploaderName} · {file.uploaderEmail} · {file.detectedMime} · {bytes(file.sizeBytes)}</small></div><span className="file-date">{new Date(file.uploadedAt).toLocaleDateString()}</span><span className="muted">{file.downloadCount} downloads</span><button className="secondary" disabled={busy} onClick={() => void share(file.id)}>Share</button></div>)}{next && <button className="secondary" onClick={() => void refresh(next)}>Load more</button>}</div>}</section>
}
