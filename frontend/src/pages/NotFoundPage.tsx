import { Link } from 'react-router-dom'

export function NotFoundPage() {
  return (
    <main className="public-shell">
      <section className="state-card" aria-labelledby="not-found-title">
        <span className="eyebrow">Erro 404</span>
        <h1 id="not-found-title">Página não encontrada</h1>
        <p>O endereço informado não corresponde a uma rota deste portal.</p>
        <Link className="button" to="/">Voltar ao portal</Link>
      </section>
    </main>
  )
}
