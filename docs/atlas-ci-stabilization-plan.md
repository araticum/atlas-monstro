# Atlas CI stabilization plan

1. **Mover o workflow de testes para GitHub-hosted runners**  
   Padronizar os jobs do CI principal em `runs-on: ubuntu-latest` para eliminar a instabilidade operacional dos self-hosted e voltar para uma base previsível.

2. **Remover hacks específicos de self-hosted**  
   Apagar passos de permissões, `chown`, correções de workspace/temp e workarounds ligados a `_work`, `/__w/_temp` ou root-owned workspace que só existiam por causa do host local.

3. **Padronizar acesso a services conforme o tipo de job**  
   Jobs host-based continuam falando com services via `localhost` + `ports:` quando necessário; jobs em `container:` usam alias de service da própria rede do Actions (ex.: `dex`) para evitar acoplamento a comportamento de runner local.
