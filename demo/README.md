# Demo - Ability System v4.6 + Multiplayer RBAC

Demonstração completa do **Ability System** com suporte a **Multiplayer, RBAC (Role-Based Access Control), e Admin Tools** para Godot 4.6+.

## 🚀 Quick Start

### Para Jogar (Novo!)
- Abra o projeto
- Pressione **F5** para executar
- Digite seu nome de jogador
- Escolha entre **Singleplayer** ou **Multiplayer**
- Jogue!

### Para Testar Sistemas Específicos
- **`demo.tscn`** - Cena principal da demo original
- **`player/player.tscn`** - Cena isolada do player
- **`ui/hud.tscn`** - Interface de usuário
- **`tests/test_rbac_scene.tscn`** - Testes de RBAC

Pressione **Play (F5)** para executar.

### Controles

| Tecla | Ação |
|-------|------|
| **WASD** | Movimentação (Idle/Walk/Run) |
| **Space** | Pulo (Jump/Fall) |
| **Mouse Click** | Ataque/Combo |
| **Shift + Click** | Dash Attack |
| **T** | Chat (Multiplayer) |
| **ESC** | Menu de Pausa (Parar jogo) |

## 📚 Documentação

### Novo! Sistema Multiplayer & RBAC
- **[QUICK_START_GUIDE.md](QUICK_START_GUIDE.md)** - Guia de início rápido para jogar
- **[RBAC_SYSTEM_GUIDE.md](RBAC_SYSTEM_GUIDE.md)** - Documentação técnica completa do sistema RBAC
- **[IMPLEMENTATION_CHECKLIST.md](IMPLEMENTATION_CHECKLIST.md)** - Status de implementação e checklist de testes
- **[FINAL_SUMMARY.md](FINAL_SUMMARY.md)** - Resumo final da implementação

### Ability System Original
- **[ABILITY_SYSTEM_DEMO.md](ABILITY_SYSTEM_DEMO.md)** - Guia completo da arquitetura e implementação
- **[Framework Documentation](../src/doc_classes/)** - Documentação técnica da API
- **[Business Rules (Português)](../BUSINESS_RULES.pt.md)** - Regras de design do framework
- **[Implementation Plan](../IMPLEMENTATION_PLAN.md)** - Plano de implementação e roadmap

## ✨ Recursos Novos - Multiplayer & RBAC

### Multiplayer
- ✅ Modo Singleplayer (local)
- ✅ Modo Multiplayer Server (host)
- ✅ Modo Multiplayer Client (join)
- ✅ Abrir mundo para LAN (Open to LAN)
- ✅ Gerenciamento de jogadores
- ✅ Sistema de chat multiplayer

### Admin & Controle de Acesso
- ✅ Sistema RBAC (Admin, Moderator, User)
- ✅ 8 Comandos de Admin (kick, ban, mute, broadcast, etc)
- ✅ Painel de Admin em-game
- ✅ Sistema de bans com expiração automática
- ✅ Terminal CLI para servidor dedicado
- ✅ Sistema de logging centralizado

## 📂 Estrutura

```
demo/
├── ui/                          ← Interface (NOVO & EXISTENTE)
│   ├── user_registration.tscn/.gd    ← NOVO: Registro de nome
│   ├── main_menu.tscn/.gd            ← NOVO: Menu principal
│   ├── admin_panel.tscn/.gd          ← NOVO: Painel de admin
│   ├── pause_menu.gd                 ← NOVO: Menu de pausa
│   ├── chat_system.tscn/.gd          ← NOVO: Sistema de chat
│   ├── hud.tscn/.gd                  ← Existing
│   └── ...
├── autoload/                    ← Global managers
│   ├── RBAC.gd                       ← NOVO: Controle de acesso
│   ├── AdminCommand.gd               ← NOVO: Processador de comandos
│   ├── AdminTerminal.gd              ← NOVO: CLI para servidor
│   ├── ServerLogger.gd               ← NOVO: Sistema de logging
│   ├── SystemVerification.gd         ← NOVO: Verificação do sistema
│   ├── MultiplayerGameManager.gd     ← NOVO: Gerenciador multiplayer
│   ├── GameData.gd                   ← NOVO: Estado global
│   └── GameMachine.gd                ← Existing
├── player/                      ← Player logic & assets
│   ├── player.tscn/.gd
│   └── resources/
├── scenes/                      ← Level scenes
│   ├── level.tscn/.gd
│   └── demo.tscn
├── server/                      ← NOVO: Servidor dedicado
│   └── server.tscn/.gd
├── tests/                       ← NOVO: Suite de testes
│   ├── test_rbac_system.gd
│   └── test_rbac_scene.tscn
├── resources/                   ← Ability System data
│   ├── abilities/
│   ├── attributes/
│   ├── effects/
│   ├── cues/
│   └── container/
└── project.godot                ← ATUALIZADO: Novos autoloads
```

## 🎮 Gameplay

Esta demo demonstra:

- ✅ Movimentação dinâmica (Idle → Walk → Run → Jump → Fall)
- ✅ Sistema de ataque com combos (3x Light → Heavy)
- ✅ Recursos (Stamina) com regeneração
- ✅ Tags dinâmicas para rastreamento de estado
- ✅ UI reativa sincronizada com estado do player
- ✅ Animações sincronizadas com estado

## 🔧 Técnicas Utilizadas

- **ASComponent**: Hub central de lógica do actor
- **Abilities**: Habilidades ativadas por input/eventos
- **Attributes**: Sistema de valores (health, stamina, etc)
- **Tags**: Rastreamento de estado (NAME, CONDITIONAL, EVENT)
- **Signals**: Atualização reativa de UI via `tag_changed`
- **Signal Callbacks**: HUD conectado a sinais, sem polling

## 🐛 Debugging

Logs habilitados para verificação:

```gdscript
# Em demo/player/player.gd
[State] Activating motion.idle
[State] Activating motion.walk
[Tag Color] Player color changed to: blue

# Em demo/ui/hud.gd
[HUD] tag_changed: ability.combo1 added
[HUD] State updated: idle → walk
[HUD] Weapon updated: unarmed
```

Verifique a **Output Console** do Godot durante execução.

## ⚡ Performance

- **Zero Polling**: UI atualiza apenas quando estado realmente muda
- **Signal-Based**: Uso de Godot signals para reatividade
- **Efficient Queries**: Busca de atributos em O(1) com HashMap

## 🎮 Compilar Executáveis

### Automático (GitHub Actions)
Os executáveis são compilados automaticamente ao fazer push para `master`, `tests` ou `dev`:

```
Windows:    demo_windows_client.exe, demo_windows_server.exe
Linux:      demo_linux_client, demo_linux_server
macOS:      demo_macos_client, demo_macos_server
Android:    demo_android_client.apk, demo_android_server.apk
```

Veja `.github/workflows/build_demo_executables.yml`

### Manual (Godot 4.6.1+)
```bash
# Baixar Godot 4.6.1 e templates
godot --path . --export-release "Windows Desktop" ../demo_windows_client.exe
godot --path . --export-release "Linux/X11" ../demo_linux_client
```

## 🧪 Testar Sistema RBAC

Abra a cena de teste:
```
demo/tests/test_rbac_scene.tscn
```

Pressione F5 para executar testes automatizados. Verifique o console de saída do Godot.

## 📖 Próximas Etapas

Para expandir a demo:

### Sistema Multiplayer
1. **Adicionar Whitelist**: Permitir apenas jogadores específicos
2. **Persistência de Bans**: Salvar lista de bans
3. **Audit Log**: Registrar todas as ações de admin
4. **Rate Limiting**: Prevenir spam de comandos

### Ability System Original
1. **Adicionar Habilidades**: Crie `.tres` em `resources/abilities/`
2. **Adicionar Efeitos**: Crie `.tres` em `resources/effects/`
3. **Adicionar Inimigos**: Use ASComponent para IA
4. **Melhorar UI**: Expanda `hud.gd` com novos painéis
5. **Adicionar Persistência**: Use ASStateSnapshot para save/load

## 📞 Suporte

### Multiplayer & RBAC
Consulte a documentação nova:
- [QUICK_START_GUIDE.md](QUICK_START_GUIDE.md) - Como jogar
- [RBAC_SYSTEM_GUIDE.md](RBAC_SYSTEM_GUIDE.md) - Documentação técnica
- [IMPLEMENTATION_CHECKLIST.md](IMPLEMENTATION_CHECKLIST.md) - Checklist de features

### Ability System Original
Consulte:
- [ABILITY_SYSTEM_DEMO.md](ABILITY_SYSTEM_DEMO.md) - Arquitetura completa
- [../IMPLEMENTATION_PLAN.md](../IMPLEMENTATION_PLAN.md) - Roadmap do projeto
- [../BUSINESS_RULES.pt.md](../BUSINESS_RULES.pt.md) - Regras de design

---

**Versão**: 4.6 with Multiplayer RBAC | **Godot**: 4.6+ | **Linguagem**: GDScript
