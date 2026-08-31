import { FolderOpen, Link2, Pencil, Trash2, X } from 'lucide-react'
import { Folder } from './api'
import './folder-context.css'

type Props = { folder: Folder; x: number; y: number; onOpen: () => void; onRename: () => void; onShare: () => void; onDelete: () => void; onClose: () => void }

export default function FolderContextMenu({ folder, x, y, onOpen, onRename, onShare, onDelete, onClose }: Props) {
  return <div className="folder-context-backdrop" onClick={onClose} onContextMenu={e => { e.preventDefault(); onClose() }}><div className="folder-context-menu" style={{ left: Math.min(x, window.innerWidth - 230), top: Math.min(y, window.innerHeight - 250) }} onClick={e => e.stopPropagation()}><div className="folder-context-title"><strong>{folder.name}</strong><button className="icon-button" onClick={onClose} aria-label="Close folder menu"><X /></button></div><button onClick={onOpen}><FolderOpen /> Open</button><button onClick={onRename}><Pencil /> Rename</button><button onClick={onShare}><Link2 /> Share</button><button className="danger-item" onClick={onDelete}><Trash2 /> Delete</button></div></div>
}
