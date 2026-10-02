import { AlertMessage } from '../components/feedback/AlertMessage'
import { PageHeader } from '../components/layout/PageHeader'

export function RoutePlaceholderPage({ title, description, backendReady = true }: { title: string; description: string; backendReady?: boolean }) {
  return (
    <section aria-labelledby="route-title">
      <PageHeader eyebrow="Estrutura de rota" title={title} description={description} titleId="route-title" />
      <AlertMessage>
        {backendReady
          ? 'O endpoint correspondente já existe, mas esta tela ainda não busca nem apresenta dados. O fluxo será implementado na próxima etapa.'
          : 'Este fluxo depende de um endpoint ainda planejado. Nenhuma resposta simulada será apresentada.'}
      </AlertMessage>
    </section>
  )
}
