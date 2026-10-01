import { Link } from 'react-router-dom'

export function NotFoundPage() {
  return (
    <main className="page-shell">
      <section className="status-card" aria-labelledby="not-found-title">
        <span className="eyebrow">Erro 404</span>
        <h1 id="not-found-title">Página não encontrada</h1>
        <p>O endereço informado não existe nesta etapa do portal.</p>
        <Link className="text-link" to="/">
          Voltar ao início
        </Link>
      </section>
    </main>
  )
}

