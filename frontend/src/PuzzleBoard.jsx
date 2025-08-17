import { useState, useEffect } from "react";

const goalState = [1, 2, 3, 4, 5, 6, 7, 8, 0];

export default function PuzzleBoard() {
  const [board, setBoard] = useState(goalState);
  const [clicks, setClicks] = useState(0);
  const [winner, setWinner] = useState(false);

  // Verifica si ya se armó la matriz
  useEffect(() => {
    if (JSON.stringify(board) === JSON.stringify(goalState)) {
      setWinner(true);
    } else {
      setWinner(false);
    }
  }, [board]);

  // Intenta mover una ficha
  const handleClick = (index) => {
    const emptyIndex = board.indexOf(0);
    const row = Math.floor(index / 3);
    const col = index % 3;
    const emptyRow = Math.floor(emptyIndex / 3);
    const emptyCol = emptyIndex % 3;

    // Solo mueve si está al lado del espacio vacío
    if (
      (row === emptyRow && Math.abs(col - emptyCol) === 1) ||
      (col === emptyCol && Math.abs(row - emptyRow) === 1)
    ) {
      const newBoard = [...board];
      [newBoard[index], newBoard[emptyIndex]] = [
        newBoard[emptyIndex],
        newBoard[index],
      ];
      setBoard(newBoard);
      setClicks((c) => c + 1);
    }
  };

  return (
    <div className="flex flex-col items-center gap-4">
      <h1 className="text-2xl font-bold">8 Puzzle</h1>
      <div className="grid grid-cols-3 gap-2">
        {board.map((tile, i) => (
          <button
            key={i}
            onClick={() => handleClick(i)}
            className={`w-16 h-16 flex items-center justify-center text-lg font-bold rounded 
              ${tile === 0 ? "bg-gray-300" : "bg-blue-500 text-white hover:bg-blue-600"}`}
          >
            {tile !== 0 ? tile : ""}
          </button>
        ))}
      </div>
      <p className="text-lg">Clics: {clicks}</p>
      {winner && (
        <p className="text-green-600 font-bold text-xl animate-bounce">
          🎉 ¡Ganador! 🎉
        </p>
      )}
    </div>
  );
}
