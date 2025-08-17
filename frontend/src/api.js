const API = 'http://localhost:8080/api'

export async function apiInit() {
  const res = await fetch(`${API}/init`)
  if (!res.ok) throw new Error('Init failed')
  return res.json()
}

export async function apiShuffle(steps) {
  const res = await fetch(`${API}/shuffle`, {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({ steps })
  })
  if (!res.ok) throw new Error('Shuffle failed')
  return res.json()
}

export async function apiSolve(start) {
  const res = await fetch(`${API}/solve`, {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({ start })
  })
  if (!res.ok) throw new Error('Solve failed')
  return res.json()
}
