import { useEffect, useRef, useState } from 'react'
import { Move, X } from 'lucide-react'
import './AvatarCropper.css'

// The chosen file remains in memory only. Canvas exports a cropped,
// optimized square and never uploads the original photo.
export default function AvatarCropper({ file, onCancel, onSave, saving }) {
    const [source, setSource] = useState(null)
    const [zoom, setZoom] = useState(1)
    const [position, setPosition] = useState({ x: 0, y: 0 })
    const [error, setError] = useState('')
    const pointer = useRef(null)
    const canvas = useRef(null)

    useEffect(() => {
        if (!file) return undefined
        const url = URL.createObjectURL(file)
        const img = new Image()
        let active = true
        img.onload = () => {
            if (!active) return
            if (img.naturalWidth > 6000 || img.naturalHeight > 6000) {
                setError('Choose a picture no larger than 6000 × 6000 pixels.')
            } else {
                setSource(img)
                setZoom(1)
                setPosition({ x: 0, y: 0 })
                setError('')
            }
        }
        img.onerror = () => { if (active) setError('This picture could not be opened. Try a PNG or JPEG.') }
        img.src = url
        return () => { active = false; URL.revokeObjectURL(url) }
    }, [file])

    useEffect(() => {
        if (!source || !canvas.current) return
        const ctx = canvas.current.getContext('2d')
        if (!ctx) return
        const side = 512
        const base = Math.max(side / source.naturalWidth, side / source.naturalHeight)
        const width = source.naturalWidth * base * zoom
        const height = source.naturalHeight * base * zoom
        ctx.clearRect(0, 0, side, side)
        ctx.drawImage(source, (side - width) / 2 + position.x, (side - height) / 2 + position.y, width, height)
    }, [source, zoom, position])

    function clamp(next, scale = zoom) {
        if (!source) return { x: 0, y: 0 }
        const base = Math.max(512 / source.naturalWidth, 512 / source.naturalHeight)
        const halfX = Math.max(0, (source.naturalWidth * base * scale - 512) / 2)
        const halfY = Math.max(0, (source.naturalHeight * base * scale - 512) / 2)
        return { x: Math.max(-halfX, Math.min(halfX, next.x)), y: Math.max(-halfY, Math.min(halfY, next.y)) }
    }
    function pointerDown(e) {
        if (!source || saving) return
        e.currentTarget.setPointerCapture(e.pointerId)
        pointer.current = { x: e.clientX, y: e.clientY, position }
    }
    function pointerMove(e) {
        if (!pointer.current) return
        const rect = e.currentTarget.getBoundingClientRect()
        const ratio = 512 / rect.width
        setPosition(clamp({
            x: pointer.current.position.x + (e.clientX - pointer.current.x) * ratio,
            y: pointer.current.position.y + (e.clientY - pointer.current.y) * ratio,
        }))
    }
    function pointerUp(e) {
        pointer.current = null
        if (e.currentTarget.hasPointerCapture(e.pointerId)) e.currentTarget.releasePointerCapture(e.pointerId)
    }
    function zoomTo(e) {
        const next = Number(e.target.value)
        setZoom(next)
        setPosition((current) => clamp(current, next))
    }
    async function save() {
        if (!canvas.current || !source || saving) return
        setError('')
        try {
            const blob = await new Promise((resolve) => canvas.current.toBlob(resolve, 'image/jpeg', 0.86))
            if (!blob) throw new Error('Could not prepare this picture.')
            if (blob.size > 512 * 1024) throw new Error('The cropped picture is too large. Try another photo.')
            await onSave(new File([blob], 'avatar.jpg', { type: 'image/jpeg' }))
        } catch (err) {
            setError(err.message || 'Could not prepare this picture.')
        }
    }

    return <div className="avatar-cropper-backdrop" role="presentation" onMouseDown={(e) => {
        if (e.target === e.currentTarget && !saving) onCancel()
    }}>
        <section className="avatar-cropper-dialog" role="dialog" aria-modal="true" aria-labelledby="avatar-cropper-title">
            <div className="avatar-cropper-heading">
                <div><h2 id="avatar-cropper-title">Adjust your picture</h2><p>Drag to position, then adjust the zoom.</p></div>
                <button type="button" aria-label="Close picture editor" onClick={onCancel} disabled={saving}><X size={20} /></button>
            </div>
            {source ? <>
                <div className="avatar-cropper-stage" onPointerDown={pointerDown}
                    onPointerMove={pointerMove} onPointerUp={pointerUp}
                    onPointerCancel={pointerUp} aria-label="Drag to reposition your picture">
                    <canvas width="512" height="512" ref={canvas} />
                    <div className="avatar-cropper-circle" aria-hidden="true" />
                </div>
                <label className="avatar-cropper-zoom" htmlFor="avatar-cropper-slider"><Move size={17} /> Zoom
                    <input id="avatar-cropper-slider" type="range" min="1" max="3" step="0.02" value={zoom}
                        onChange={zoomTo} disabled={saving} />
                </label>
            </> : !error ? <p role="status">Opening picture…</p> : null}
            {error && <p className="avatar-cropper-error" role="alert">{error}</p>}
            <div className="avatar-cropper-actions">
                <button type="button" onClick={onCancel} disabled={saving}>Cancel</button>
                <button type="button" onClick={save} disabled={saving || !source}>
                    {saving ? 'Saving…' : 'Save picture'}
                </button>
            </div>
        </section>
    </div>
}
