// frontend/src/App.jsx
import { useEffect, useMemo, useRef, useState } from 'react'
import { apiInit, apiShuffle, apiSolve } from './api.js'

export default function App() {
  const [state, setState] = useState([1,2,3,4,5,6,7,8,0])
  const [solution, setSolution] = useState([])
  const [actions, setActions] = useState([])
  const [stepIndex, setStepIndex] = useState(0)
  const [nShuffle, setNShuffle] = useState(20)
  const [busy, setBusy] = useState(false)
  const [clicks, setClicks] = useState(0)  
  const animRef = useRef(null)

  useEffect(() => {
    apiInit().then(d => setState(d.state)).catch(()=>{})
  }, [])

  const isSolved = useMemo(
    () => JSON.stringify(state) === JSON.stringify([1,2,3,4,5,6,7,8,0]),
    [state]
  )

  function resetSolution(){
    setSolution([]); setActions([]); setStepIndex(0)
  }

  function clearAnimation(){
    if (animRef.current){
      clearInterval(animRef.current)
      animRef.current = null
    }
  }

  async function onInit() {
    clearAnimation()
    const d = await apiInit()
    setState(d.state)
    resetSolution()
    setClicks(0) 
  }

  async function onShuffle() {
    clearAnimation()
    setBusy(true)
    try {
      const d = await apiShuffle(Number(nShuffle||20))
      setState(d.state)
      resetSolution()
      setClicks(0) 
    } finally { setBusy(false) }
  }

  async function onSolve(autoPlay=false) {
    setBusy(true)
    try {
      const d = await apiSolve(state)
      setSolution(d.moves || [])
      setActions(d.actions || [])
      setStepIndex(0)
      if (autoPlay && (d.moves?.length ?? 0) > 0) {
        playAnimation(d.moves)
      }
    } finally { setBusy(false) }
  }

  function playAnimation(moves) {
    clearAnimation()
    let i = 0
    animRef.current = setInterval(() => {
      if (i >= moves.length) {
        clearAnimation()
        return
      }
      setState(moves[i])
      i++
    }, 500)
  }

  // --------- mover fichas con clic ---------
  function isAdjacent(idxA, idxB) {
    const rA = Math.floor(idxA / 3), cA = idxA % 3
    const rB = Math.floor(idxB / 3), cB = idxB % 3
    return (rA === rB && Math.abs(cA - cB) === 1) || (cA === cB && Math.abs(rA - rB) === 1)
  }

  function handleTileClick(idx) {
    const blank = state.indexOf(0)
    if (idx === blank) return
    if (isAdjacent(idx, blank)) {
      const next = [...state]
      next[blank] = state[idx]
      next[idx] = 0
      setState(next)
      setClicks(c => c + 1) 
      clearAnimation()
      resetSolution()
    }
  }

  const moveableSet = useMemo(() => {
    const blank = state.indexOf(0)
    return new Set(
      [blank-3, blank+3, blank-1, blank+1].filter(i => i>=0 && i<9)
        .filter(i => isAdjacent(i, blank))
    )
  }, [state])
  // -----------------------------------------

  function nextStep(){
    if (solution.length === 0) return
    const next = Math.min(stepIndex + 1, solution.length - 1)
    setStepIndex(next)
    setState(solution[next])
  }

  function prevStep(){
    if (solution.length === 0) return
    const prev = Math.max(stepIndex - 1, 0)
    setStepIndex(prev)
    setState(solution[prev])
  }

  return (
    <div className="container">
      <h1>8 Puzzle</h1>

      <div className="status">
        <p><strong>Clics:</strong> {clicks}</p> 
        {isSolved && <p className="winner"> ¡Ganador! </p>} 
      </div>

      <div className="panel">
        <Grid state={state} onTileClick={handleTileClick} moveableSet={moveableSet} />

        <div className="controls">
          <div className="row">
            <button onClick={onInit} disabled={busy}>Iniciar</button>
            <input type="number" min="1" max="200" value={nShuffle} onChange={e=>setNShuffle(e.target.value)} />
            <button onClick={onShuffle} disabled={busy}>Desordenar</button>
          </div>

          <div className="row">
            <button onClick={()=>onSolve(true)} disabled={busy || isSolved}>Resolver (auto)</button>
            <button onClick={()=>onSolve(false)} disabled={busy || isSolved}>Resolver (preparar pasos)</button>
          </div>

          <div className="row">
            <button onClick={prevStep} disabled={solution.length===0 || stepIndex===0 || busy}>Paso anterior</button>
            <button onClick={nextStep} disabled={solution.length===0 || stepIndex===solution.length-1 || busy}>Siguiente paso</button>
            {solution.length>0 && <span className="badge">Pasos: <strong>{solution.length-1}</strong></span>}
          </div>

          <div className="footer">
            Estado actual: <code>[{state.join(', ')}]</code>
          </div>
        </div>
      </div>
    </div>
  )
}

function Grid({ state, onTileClick, moveableSet }){
  return (
    <div className="grid">
      {state.map((n, idx)=> {
        const cls = [
          'tile',
          n===0 ? 'zero' : '',
          moveableSet?.has(idx) && n!==0 ? 'moveable' : ''
        ].filter(Boolean).join(' ')
        return (
          <div
            key={idx}
            className={cls}
            onClick={() => n!==0 && onTileClick(idx)}
          >
            {n!==0 ? n : ''}
          </div>
        )
      })}
    </div>
  )
}
