import { useEffect, useRef, useState } from 'react'
import { createRoot } from 'react-dom/client'
import { Clipboard, CloudUpload, Download, Eye, FileText, Folder, HardDrive, Link2, LogOut, Plus, Search, ShieldCheck, Trash2, X } from 'lucide-react'
import { api, FileDetail, FileItem, Folder as FolderType, register, Stats, User } from './api'
import AdminPanel from './AdminPanel'
import ShareDialog from './ShareDialog'
import StorageStatsVisual from './StorageStatsVisual'
import FolderContextMenu from './FolderContextMenu'

type View = 'all' | 'folders' | 'uploads' | 'admin'
const bytes = (n: number) => n < 1024 ? `${n} B` : n < 1048576 ? `${(n / 1024).toFixed(1)} KB` : `${(n / 1048576).toFixed(1)} MB`

export default function VaultWorkspace() {
  const [user, setUser] = useState<User | null>(null)
  const [email, setEmail] = useState('')
  const [displayName, setDisplayName] = useState('')
  const [registerMode, setRegisterMode] = useState(false)
  const [password, setPassword] = useState('')
  const [csrf, setCsrf] = useState('')
  const [view, setView] = useState<View>('all')
  const [folderId, setFolderId] = useState('')
  const [query, setQuery] = useState('')
  const [mime, setMime] = useState('')
  const [minSize, setMinSize] = useState('')
  const [maxSize, setMaxSize] = useState('')
  const [afterDate, setAfterDate] = useState('')
  const [beforeDate, setBeforeDate] = useState('')
  const [uploader, setUploader] = useState('')
  const [tag, setTag] = useState('')
  const [files, setFiles] = useState<FileItem[]>([])
  const [selectedFiles, setSelectedFiles] = useState<Set<string>>(new Set())
  const [folders, setFolders] = useState<FolderType[]>([])
  const [stats, setStats] = useState<Stats | null>(null)
  const [selected, setSelected] = useState<FileItem | null>(null)
  const [detail, setDetail] = useState<FileDetail | null>(null)
  const [pendingFolderId, setPendingFolderId] = useState('')
  const [shareOpen, setShareOpen] = useState(false)
  const [shareEmail, setShareEmail] = useState('')
  const [permission, setPermission] = useState('view')
  const [publicLink, setPublicLink] = useState('')
  const [copied, setCopied] = useState(false)
  const [folderShareLink, setFolderShareLink] = useState('')
  const [folderCopied, setFolderCopied] = useState(false)
  const [previewOpen, setPreviewOpen] = useState(false)
  const [error, setError] = useState('')
  const [uploading, setUploading] = useState(false)
  const [dragging, setDragging] = useState(false)
  const [shareTarget, setShareTarget] = useState<{ kind: 'file' | 'folder'; id: string; name: string } | null>(null)
  const [folderContext, setFolderContext] = useState<{ folder: FolderType; x: number; y: number } | null>(null)
  const loading = useRef(false)
  const uploadRef = useRef<(incoming: File[]) => void>(() => undefined)

  const load = async () => {
    if (loading.current) return
    loading.current = true
    try {
      let token = csrf
      if (!token) { const c = await api.csrf(); token = c.csrfToken; setCsrf(token) }
      const params = new URLSearchParams()
      if (folderId) params.set('folderId', folderId)
      if (query.trim()) params.set('filename', query.trim())
      if (mime.trim()) params.set('mime', mime.trim())
      if (minSize.trim()) params.set('minSizeBytes', minSize.trim())
      if (maxSize.trim()) params.set('maxSizeBytes', maxSize.trim())
      if (afterDate) params.set('uploadedAfter', new Date(`${afterDate}T00:00:00Z`).toISOString())
      if (beforeDate) params.set('uploadedBefore', new Date(`${beforeDate}T23:59:59Z`).toISOString())
      if (uploader.trim()) params.set('uploaderName', uploader.trim())
      if (tag.trim()) params.set('tag', tag.trim())
      const suffix = params.toString() ? `?${params.toString()}` : ''
      const [s, f, d] = await Promise.all([api.stats(), view === 'folders' && !folderId ? Promise.resolve({ items: [] as FileItem[] }) : api.files(suffix), api.folders(folderId)])
      setStats(s); setFiles(f.items); setFolders(d); setError('')
    } catch (e) { setError(e instanceof Error ? e.message : 'Unable to load vault') }
    finally { loading.current = false }
  }

  useEffect(() => {
    if (!user) return
    const timer = window.setTimeout(() => void load(), 250)
    return () => window.clearTimeout(timer)
  }, [user, view, folderId, query, mime, minSize, maxSize, afterDate, beforeDate, uploader, tag])
  useEffect(() => {
    document.body.dataset.vaultView = view
    const mimeSelect = document.querySelector<HTMLSelectElement>('.filter-bar .mime-filter')
    if (mimeSelect) mimeSelect.value = mime
  }, [view, user])
  useEffect(() => {
    const isFileDrag = (event: DragEvent) => event.dataTransfer?.types.includes('Files') || Array.from(event.dataTransfer?.items ?? []).some(item => item.kind === 'file')
    const preventBrowserNavigation = (event: DragEvent) => {
      if (!isFileDrag(event)) return
      event.preventDefault()
      if (event.type === 'drop' && (view !== 'folders' || Boolean(folderId))) {
        event.stopPropagation(); setDragging(false)
        uploadRef.current(Array.from(event.dataTransfer?.files ?? []))
      }
    }
    window.addEventListener('dragover', preventBrowserNavigation, true)
    window.addEventListener('drop', preventBrowserNavigation, true)
    return () => { window.removeEventListener('dragover', preventBrowserNavigation, true); window.removeEventListener('drop', preventBrowserNavigation, true) }
  }, [view, folderId])
  useEffect(() => { api.me().then(setUser).catch(() => setUser(null)) }, [])
  useEffect(() => {
    if (!user) return
    const cards = Array.from(document.querySelectorAll<HTMLElement>('.folder-card'))
    const handlers = cards.map(card => {
      const handler = () => {
        const button = card.querySelector<HTMLButtonElement>('.folder-open')
        if (!button) return
        const folder = folders.find(item => item.name === button.textContent?.trim())
        if (folder) window.history.pushState({ vault: true, folderId: folder.id }, '', `#folder=${encodeURIComponent(folder.id)}`)
      }
      card.addEventListener('click', handler, true)
      return { card, handler }
    })
    const onPopState = () => {
      if (window.location.hash.startsWith('#folder=')) {
        const id = typeof window.history.state?.folderId === 'string' ? window.history.state.folderId : decodeURIComponent(window.location.hash.slice('#folder='.length))
        setView('folders'); setFolderId(id); setSelectedFiles(new Set()); return
      }
      setFolderId(''); setView('all'); setSelectedFiles(new Set()); setFolderContext(null)
    }
    const onBackClick = (event: MouseEvent) => {
      const target = event.target instanceof HTMLElement ? event.target.closest('button') : null
      if (!target || !window.location.hash.startsWith('#folder=')) return
      if (target.textContent?.trim() !== 'Back to folders' && target.textContent?.trim() !== 'Back to home') return
      event.preventDefault(); event.stopImmediatePropagation(); window.history.back()
    }
    document.addEventListener('click', onBackClick, true)
    window.addEventListener('popstate', onPopState)
    return () => { handlers.forEach(({ card, handler }) => card.removeEventListener('click', handler, true)); document.removeEventListener('click', onBackClick, true); window.removeEventListener('popstate', onPopState) }
  }, [user, folders])

  const signIn = async (e: React.FormEvent) => { e.preventDefault(); setError(''); try { const u = registerMode ? await register(email, displayName, password) : await api.login(email, password); const c = await api.csrf(); setCsrf(c.csrfToken); setUser(u) } catch (e) { setError(e instanceof Error ? e.message : registerMode ? 'Unable to create your account' : 'Unable to sign in') } }
  const upload = async (incoming: File[]) => {
    if (!incoming.length) return
    setUploading(true); const errors: string[] = []
    try { const fresh = await api.csrf(); setCsrf(fresh.csrfToken); for (const file of incoming) { try { const r = await api.upload(file, fresh.csrfToken); const bad = r.results.find(x => x.status === 'rejected'); if (bad) errors.push(`${bad.filename}: ${bad.error?.message ?? 'Upload rejected'}`) } catch (e) { errors.push(e instanceof Error ? e.message : 'Upload failed') } } await load() }
    catch (e) { errors.push(e instanceof Error ? e.message : 'Upload failed') }
    finally { setUploading(false); if (errors.length) setError(errors.join(' ')) }
  }
  uploadRef.current = incoming => { void upload(incoming) }
  const showDetails = async (file: FileItem) => { setSelected(file); setDetail(null); setPendingFolderId(file.folderId ?? ''); setShareOpen(false); setPublicLink(''); setCopied(false); setPreviewOpen(false); try { setDetail(await api.detail(file.id)) } catch (e) { setError(e instanceof Error ? e.message : 'Unable to load file details') } }
  const createFolder = async () => { const name = window.prompt('Folder name'); if (!name) return; try { await api.createFolder(name, folderId, csrf); await load() } catch (e) { setError(e instanceof Error ? e.message : 'Unable to create folder') } }
  const renameFolder = async (folder: FolderType) => { const name = window.prompt('Rename folder', folder.name); if (!name || name.trim() === folder.name) return; try { await api.renameFolder(folder.id, name, csrf); await load() } catch (e) { setError(e instanceof Error ? e.message : 'Unable to rename folder') } }
  const deleteFolder = async (folder: FolderType) => { if (!window.confirm(`Delete the empty folder “${folder.name}”?`)) return; try { await api.deleteFolder(folder.id, csrf); await load() } catch (e) { setError(e instanceof Error ? e.message : 'Unable to delete folder') } }
  const shareFolder = (folder: FolderType) => setShareTarget({ kind: 'folder', id: folder.id, name: folder.name })
  const move = async () => { if (!selected) return; try { await api.moveFile(selected.id, pendingFolderId, csrf); const updated = await api.detail(selected.id); setDetail(updated); setSelected({ ...selected, folderId: pendingFolderId || undefined }); await load() } catch (e) { setError(e instanceof Error ? e.message : 'Unable to move file') } }
  const bulkDelete = async () => { if (!selectedFiles.size || !window.confirm(`Delete ${selectedFiles.size} selected file${selectedFiles.size === 1 ? '' : 's'}?`)) return; const errors: string[] = []; for (const id of selectedFiles) { try { await api.deleteFile(id, csrf) } catch (e) { errors.push(e instanceof Error ? e.message : 'Unable to delete selected file') } } setSelectedFiles(new Set()); await load(); if (errors.length) setError(errors.join(' ')) }
  const sharePublic = async () => { if (!selected) return; try { const r = await api.publicShare(selected.id, csrf); setPublicLink(`${window.location.origin}/public/${r.token}`); setCopied(false) } catch (e) { setError(e instanceof Error ? e.message : 'Unable to create public link') } }
  const copyPublicLink = async () => { if (publicLink) { await navigator.clipboard.writeText(publicLink); setCopied(true) } }
  const copyFolderLink = async () => { if (folderShareLink) { await navigator.clipboard.writeText(folderShareLink); setFolderCopied(true) } }
  const shareDirect = async () => { if (!selected || !shareEmail) return; try { await api.directShare(selected.id, shareEmail, permission, csrf); setError(''); setShareOpen(false) } catch (e) { setError(e instanceof Error ? e.message : 'Unable to share with user') } }
  const navigate = (next: View) => { setView(next); setFolderId(''); setQuery(''); setMime(''); setMinSize(''); setMaxSize(''); setAfterDate(''); setBeforeDate(''); setUploader(''); setTag(''); setFolderShareLink('') }

  useEffect(() => {
    if (!user) return
    const list = document.querySelector('.content .file-list')
    if (!list) return
    const rows = Array.from(list.querySelectorAll<HTMLElement>(':scope > .file-row'))
    rows.forEach((row, index) => {
      const file = files[index]
      if (!file) return
      const existing = row.querySelector<HTMLInputElement>('.file-select')
      if (existing) { existing.checked = selectedFiles.has(file.id); return }
      const checkbox = document.createElement('input')
      checkbox.type = 'checkbox'; checkbox.className = 'file-select'; checkbox.title = `Select ${file.name}`; checkbox.checked = selectedFiles.has(file.id)
      checkbox.addEventListener('click', event => event.stopPropagation())
       checkbox.addEventListener('change', () => setSelectedFiles(previous => { const next = new Set(previous); if (checkbox.checked) next.add(file.id); else next.delete(file.id); return next }))
      row.prepend(checkbox)
    })
    let actions = list.parentElement?.querySelector<HTMLElement>('.bulk-actions')
    if (!actions && selectedFiles.size) { actions = document.createElement('div'); actions.className = 'bulk-actions'; list.parentElement?.insertBefore(actions, list) }
    if (actions) { actions.innerHTML = ''; const label = document.createElement('span'); label.textContent = `${selectedFiles.size} selected`; const clear = document.createElement('button'); clear.className = 'secondary'; clear.textContent = 'Clear selection'; clear.onclick = () => setSelectedFiles(new Set()); const remove = document.createElement('button'); remove.className = 'danger-button'; remove.textContent = 'Delete selected'; remove.onclick = () => void bulkDelete(); actions.append(label, clear, remove); if (!selectedFiles.size) actions.remove() }
  }, [user, files, selectedFiles])
  useEffect(() => {
    if (user?.role !== 'admin') return
    const nav = document.querySelector('nav')
    if (!nav || nav.querySelector('.admin-nav')) return
    const button = document.createElement('button'); button.className = 'admin-nav'; button.textContent = 'Statistics & oversight'; button.onclick = () => setView('admin'); nav.append(button)
  }, [user])
  useEffect(() => {
    const header = document.querySelector<HTMLElement>('.content header')
    if (!user) return
    if (view === 'all' && !folderId) { header?.querySelector('.home-back')?.remove(); return }
    if (!header || header.querySelector('.home-back')) return
    const button = document.createElement('button'); button.className = 'secondary home-back'; button.textContent = 'Back to home'; button.onclick = () => { setView('all'); setFolderId(''); setSelectedFiles(new Set()) }
    header.append(button)
  }, [user, view, folderId])

  useEffect(() => {
    if (shareOpen && selected) setShareTarget({ kind: 'file', id: selected.id, name: selected.name })
  }, [shareOpen, selected])

  useEffect(() => {
    if (!shareTarget) return
    const host = document.createElement('div')
    document.body.append(host)
    const root = createRoot(host)
    root.render(<ShareDialog kind={shareTarget.kind} id={shareTarget.id} name={shareTarget.name} csrf={csrf} onClose={() => setShareTarget(null)} />)
    return () => { root.unmount(); host.remove() }
  }, [shareTarget, csrf])

  useEffect(() => {
    if (!folderContext) return
    const host = document.createElement('div')
    document.body.append(host)
    const root = createRoot(host)
    const folder = folderContext.folder
    root.render(<FolderContextMenu folder={folder} x={folderContext.x} y={folderContext.y} onOpen={() => { setFolderContext(null); setView('folders'); setFolderId(folder.id); setQuery('') }} onRename={() => { setFolderContext(null); void renameFolder(folder) }} onShare={() => { setFolderContext(null); shareFolder(folder) }} onDelete={() => { setFolderContext(null); void deleteFolder(folder) }} onClose={() => setFolderContext(null)} />)
    return () => { root.unmount(); host.remove() }
  }, [folderContext])

  if (user?.role === 'admin' && view === 'admin') return <AdminPanel visible onBack={() => setView('all')} />

  if (!user) return <main className="auth"><div className="auth-card"><div className="brand"><ShieldCheck /> BalkanID Vault</div><p className="eyebrow">SECURE FILE OPERATIONS</p><h1>{registerMode ? 'Create your private workspace.' : 'Your private workspace.'}</h1><p className="muted">{registerMode ? 'Set up your account and start organizing your files.' : 'Sign in to access your secure file workspace.'}</p><form onSubmit={signIn}>{registerMode && <label>Full name<input value={displayName} onChange={e => setDisplayName(e.target.value)} autoComplete="name" required /></label>}<label>Email<input value={email} onChange={e => setEmail(e.target.value)} type="email" autoComplete="email" required /></label><label>Password<input value={password} onChange={e => setPassword(e.target.value)} type="password" autoComplete={registerMode ? 'new-password' : 'current-password'} minLength={12} required /></label>{registerMode && <small className="auth-password-hint">Use at least 12 characters.</small>}{error && <p className="error">{error}</p>}<button className="primary">{registerMode ? 'Create account' : 'Sign in to vault'}</button></form><button className="auth-switch" onClick={() => { setRegisterMode(!registerMode); setError('') }}>{registerMode ? 'Already have an account? Sign in' : 'New here? Create an account'}</button></div></main>

  const title = folderId ? 'Folder contents' : view === 'folders' ? 'Folders' : view === 'uploads' ? 'Recent uploads' : 'All files'
  return <><AdminPanel visible={user.role === 'admin'} /><div className="app"><aside><div className="brand"><ShieldCheck /> BalkanID</div><div className="workspace"><span className="avatar">{user.displayName[0]}</span><div><strong>{user.displayName}</strong><small>Personal vault</small></div></div><nav><button className={view === 'all' && !folderId ? 'active' : ''} onClick={() => navigate('all')}><HardDrive /> All files</button><button className={view === 'folders' ? 'active' : ''} onClick={() => navigate('folders')}><Folder /> Folders</button><button className={view === 'uploads' ? 'active' : ''} onClick={() => navigate('uploads')}><CloudUpload /> Uploads</button></nav><div className="side-note"><strong>Private by default</strong><small>Files are scoped to your account.</small></div><button className="logout" onClick={async () => { await fetch('/api/auth/logout', { method: 'POST', headers: { 'X-CSRF-Token': csrf }, credentials: 'include' }); setUser(null) }}><LogOut /> Sign out</button></aside><main className="content"><header><div><p className="eyebrow">PERSONAL VAULT</p><h1>Good morning, {user.displayName.split(' ')[0]}.</h1><p className="muted">Everything secure, organized, and ready when you are.</p></div><div className="security"><span className="status-dot" /> Vault protected</div></header><section className="filter-bar"><select className="mime-filter" aria-label="MIME type filter" value={mime} onChange={e => setMime(e.target.value)}><option value="">All MIME types</option><option value="application/pdf">PDF</option><option value="application/zip">ZIP</option><option value="application/json">JSON</option><option value="application/octet-stream">Other application</option><option value="audio/mpeg">MP3 audio</option><option value="audio/wav">WAV audio</option><option value="image/png">PNG image</option><option value="image/jpeg">JPEG image</option><option value="image/gif">GIF image</option><option value="text/plain">Plain text</option><option value="text/csv">CSV text</option><option value="video/mp4">MP4 video</option></select><input placeholder="Min bytes" value={minSize} onChange={e => setMinSize(e.target.value)} /><input placeholder="Max bytes" value={maxSize} onChange={e => setMaxSize(e.target.value)} /><input placeholder="Uploader name" value={uploader} onChange={e => setUploader(e.target.value)} /><input type="date" aria-label="Uploaded after" value={afterDate} onChange={e => setAfterDate(e.target.value)} /><input type="date" aria-label="Uploaded before" value={beforeDate} onChange={e => setBeforeDate(e.target.value)} /><input placeholder="Tag" value={tag} onChange={e => setTag(e.target.value)} /><button className="secondary" onClick={() => { setMime(''); setMinSize(''); setMaxSize(''); setAfterDate(''); setBeforeDate(''); setUploader(''); setTag('') }}>Clear filters</button></section>{error && <p className="error banner">{error}</p>}<section className="stats"><div className="stat"><span>Total storage used</span><strong>{stats ? bytes(stats.deduplicatedBytes) : '—'}</strong><small>physical storage after deduplication</small></div><div className="stat"><span>Original usage</span><strong>{stats ? bytes(stats.originalBytes) : '—'}</strong><small>logical storage before deduplication</small></div><div className="stat"><span>Storage savings</span><strong>{stats ? bytes(stats.savingsBytes) : '—'}</strong><small>{stats ? `${stats.savingsPercent.toFixed(1)}% saved` : '—'}</small></div><div className="stat"><span>Files</span><strong>{files.length}</strong><small>{folderId ? 'in this folder' : 'visible in this view'}</small></div></section>{stats && <StorageStatsVisual stats={stats} />}<section className="toolbar"><div className="search"><Search /><input list="filename-suggestions" aria-label="Search files by filename" placeholder="Search files by filename" value={query} onChange={e => setQuery(e.target.value)} /><datalist id="filename-suggestions">{files.slice(0, 8).map(file => <option key={file.id} value={file.name}>{file.detectedMime} · {bytes(file.sizeBytes)}</option>)}</datalist></div><button className="secondary" onClick={createFolder}><Plus /> New folder</button><div className={`upload-dropzone ${dragging ? 'dragging' : ''}`} onDragEnter={e => { e.preventDefault(); setDragging(true) }} onDragOver={e => e.preventDefault()} onDragLeave={() => setDragging(false)} onDrop={e => { e.preventDefault(); setDragging(false); void upload(Array.from(e.dataTransfer.files)) }}><label className="upload"><CloudUpload /> {uploading ? 'Uploading…' : 'Drop files or upload'}<input type="file" multiple onChange={e => void upload(Array.from(e.target.files ?? []))} /></label></div></section><section className="file-panel"><div className="panel-head"><div><h2>{title}</h2><span className="muted">{files.length} shown</span></div>{folderId && <button className="secondary" onClick={() => { setFolderId(''); setView('folders') }}>Back to folders</button>}</div>{folders.length > 0 && <div className="folder-list">{folders.map(f => <div className="folder-card" key={f.id} onContextMenu={event => { event.preventDefault(); setFolderContext({ folder: f, x: event.clientX, y: event.clientY }) }}><button className="folder-open secondary" onClick={() => { setView('folders'); setFolderId(f.id); setQuery('') }}><Folder /> {f.name}</button></div>)}</div>}{folderShareLink && <div className="share-result folder-share"><small>Folder public link</small><div className="link-row"><code>{folderShareLink}</code><button className="icon-button" title="Copy folder link" onClick={() => void copyFolderLink()}>{folderCopied ? 'Copied' : <Clipboard />}</button></div></div>}{files.length === 0 ? <div className="empty"><FileText /><h3>{view === 'folders' && !folderId ? 'No folders yet' : 'No files here'}</h3><p>Upload a file or create a folder to get started.</p></div> : <div className="file-list">{files.map(file => <div className="file-row" key={file.id} onClick={() => void showDetails(file)}><div className="file-icon"><FileText /></div><div className="file-name"><strong>{file.name}</strong><small>{file.detectedMime} · {bytes(file.sizeBytes)}{file.wasDeduplicated ? ' · deduplicated' : ''}</small></div><span className="file-date">{new Date(file.uploadedAt).toLocaleDateString()}</span><button className="icon-button" title="Preview" onClick={e => { e.stopPropagation(); void showDetails(file).then(() => setPreviewOpen(true)) }}><Eye /></button><button className="icon-button" title="Download" onClick={e => { e.stopPropagation(); window.location.assign(`/api/files/${file.id}/content`) }}><Download /></button><button className="icon-button" title="Delete" onClick={async e => { e.stopPropagation(); if (confirm('Delete this file?')) { try { await api.deleteFile(file.id, csrf); await load() } catch (error) { setError(error instanceof Error ? error.message : 'Unable to delete file') } } }}><Trash2 /></button></div>)}</div>}</section>{selected && <div className="modal-backdrop" onClick={() => { setSelected(null); setDetail(null); setPreviewOpen(false) }}><section className="modal" onClick={e => e.stopPropagation()}><button className="modal-close" onClick={() => { setSelected(null); setDetail(null); setPreviewOpen(false) }}><X /></button><p className="eyebrow">FILE DETAILS</p><h2>{selected.name}</h2>{detail ? <><p className="muted">Uploader: {detail.uploaderName} ({detail.uploaderEmail})</p><p className="muted">Size: {bytes(detail.sizeBytes)} · Uploaded: {new Date(detail.uploadedAt).toLocaleString()}</p><p className="muted">Declared MIME: {detail.declaredMime} · Detected MIME: {detail.detectedMime}</p><p className="muted">{detail.wasDeduplicated ? 'Deduplicated reference' : 'Original blob'} · {detail.blobReferenceCount} reference(s) · {detail.downloadCount} downloads</p><label>Move to folder<select value={pendingFolderId} onChange={e => setPendingFolderId(e.target.value)}><option value="">Root</option>{folders.map(f => <option key={f.id} value={f.id}>{f.name}</option>)}</select></label><button className="primary move-button" disabled={pendingFolderId === (selected.folderId ?? '')} onClick={() => void move()}>Move</button>{previewOpen && (detail.detectedMime === 'application/pdf' ? <iframe className="preview-frame" title="File preview" src={`/api/files/${selected.id}/preview`} /> : detail.detectedMime.startsWith('image/') ? <img className="preview-image" alt={`Preview of ${selected.name}`} src={`/api/files/${selected.id}/preview`} /> : <p className="muted preview-note">Preview is available for PDF and image files. Use Download for this file type.</p>)}</> : <p className="muted">Loading details…</p>}<div className="modal-actions"><button className="primary" onClick={() => setPreviewOpen(!previewOpen)}><Eye /> {previewOpen ? 'Hide preview' : 'Preview'}</button><button className="secondary" onClick={() => setShareOpen(!shareOpen)}><Link2 /> Share</button></div>{shareOpen && <div className="share-result"><strong>Share settings</strong><button className="secondary" onClick={() => void sharePublic()}>Anyone with the link</button><label>Email<input value={shareEmail} onChange={e => setShareEmail(e.target.value)} placeholder="recipient@example.com" type="email" /></label><select value={permission} onChange={e => setPermission(e.target.value)}><option value="view">Can view</option><option value="download">Can download</option></select><button className="primary" onClick={() => void shareDirect()}>Share with user</button>{publicLink && <div className="public-link"><small>Public preview link</small><div className="link-row"><code>{publicLink}</code><button className="icon-button" title="Copy link" onClick={() => void copyPublicLink()}>{copied ? 'Copied' : <Clipboard />}</button></div></div>}</div>}</section></div>}</main></div></>
}
