# Rust: Linha do Tempo e Ecossistema (Arquivado / Visão Histórica)

> [!NOTE] > **Nota Histórica do Projeto:** Originalmente, o Vectora considerou Rust para seu núcleo. No entanto, em Março de 2026, o projeto migrou oficialmente para **Go** como linguagem principal do Core para maximizar a portabilidade e a velocidade de integração com o ecossistema de IDEs e ferramentas de sistema nativas. Este documento permanece como referência técnica da linguagem.

---

## Sumário

1. [Origem e Filosofia](#1-origem-e-filosofia)
2. [Os Pilares Técnicos do Rust](#2-os-pilares-técnicos-do-rust)
3. [O Sistema de Edições](#3-o-sistema-de-edições)
4. [FFI, Unsafe e Interoperabilidade com C/C++](#4-ffi-unsafe-e-interoperabilidade-com-cc)
5. [Linha do Tempo: Versões do Rust com Ecossistema Paralelo](#5-linha-do-tempo-versões-do-rust-com-ecossistema-paralelo)
6. [Runtime Assíncrono: Tokio e o Ecossistema Async](#6-runtime-assíncrono-tokio-e-o-ecossistema-async)
7. [Frameworks Web](#7-frameworks-web)
8. [CLI e TUI](#8-cli-e-tui)
9. [GUI e Desktop](#9-gui-e-desktop)
10. [Banco de Dados e Persistência](#10-banco-de-dados-e-persistência)
11. [Serialização e Formatos de Dados](#11-serialização-e-formatos-de-dados)
12. [Embedded e no_std](#12-embedded-e-no_std)
13. [WebAssembly](#13-webassembly)
14. [Observabilidade e Testes](#14-observabilidade-e-testes)
15. [Ferramentas de Desenvolvimento](#15-ferramentas-de-desenvolvimento)
16. [Comparativo: Antes vs Hoje](#16-comparativo-antes-vs-hoje)

---

## 1. Origem e Filosofia

### 2006: Graydon Hoare e o projeto pessoal

Rust foi criado por **Graydon Hoare** em 2006 como projeto pessoal enquanto trabalhava na Mozilla. A motivação original foi pragmática e quase anedótica: Hoare estava frustrado com um elevator crash em seu prédio causado por bug de memória no software embarcado. O objetivo era criar uma linguagem de sistemas que eliminasse classes inteiras de bugs — use-after-free, buffer overflow, data races — em tempo de compilação, sem garbage collector.

A Mozilla oficializou o patrocínio em **2009** e passou a integrar Rust no desenvolvimento do projeto **Servo** — um novo motor de browser engine em paralelo ao Gecko, projetado para explorar concorrência massiva que Gecko (em C++) não podia fazer com segurança.

### A filosofia fundacional

**Zero-cost abstractions** — abstrações de alto nível (iterators, closures, generics) compilam para código de máquina equivalente ao C manual. Você não paga pelo que não usa.

**Memory safety without GC** — o **borrow checker** rastreia ownership, lifetimes e aliasing de referências em tempo de compilação. Não existe garbage collector em runtime; memória é gerenciada deterministicamente via RAII (Resource Acquisition Is Initialization). Quando um valor sai de escopo, seu destrutor é chamado automaticamente.

**Fearless concurrency** — as mesmas garantias de ownership que previnem use-after-free também previnem data races. O compilador torna impossível compartilhar estado mutável entre threads sem sincronização explícita.

**Systems programming ergonomics** — Rust rejeita a dicotomia "seguro mas lento vs rápido mas perigoso". Compete com C e C++ em performance enquanto oferece segurança de memória por padrão.

### O que o Rust rejeita

Ao contrário de Go, Java ou C#, Rust não tem garbage collector — sem GC pauses, sem runtime overhead, sem incerteza sobre quando memória é liberada. Ao contrário de C++, Rust tem um modelo de ownership formal verificável pelo compilador — em vez de guidelines e best practices que precisam ser seguidos manualmente. Ao contrário de linguagens funcionais tipadas, Rust é explicitamente de systems programming — acesso direto a hardware, ponteiros, layout de memória, interoperabilidade com C.

---

## 2. Os Pilares Técnicos do Rust

### Ownership

Cada valor em Rust tem exatamente um **owner** — a variável que "possui" aquela memória. Quando o owner sai de escopo, o valor é dropped (destruído). Atribuição move o ownership por padrão para tipos não-Copy.

```rust
let s1 = String::from("hello");
let s2 = s1; // s1 foi movido para s2
// println!("{}", s1); // ERRO: s1 não é mais válido
println!("{}", s2); // OK
```

Tipos simples (`i32`, `f64`, `bool`, `char`, arrays pequenos) implementam `Copy` — são copiados em vez de movidos:

```rust
let x = 5;
let y = x; // x é copiado, ambos são válidos
println!("{} {}", x, y); // OK
```

### Borrowing e Referências

Para usar um valor sem tomar ownership, você o **empresta** via referência. Existem duas regras imutáveis que o borrow checker aplica:

1. Em qualquer momento, você pode ter **qualquer número de referências imutáveis** (`&T`) OU **exatamente uma referência mutável** (`&mut T`) — nunca ambos simultaneamente.
2. Referências devem sempre ser válidas (nunca apontar para memória liberada).

```rust
fn calculate_length(s: &String) -> usize { // referência imutável
    s.len()
} // s não é dropped aqui — não possui o valor

let mut s = String::from("hello");
let r1 = &s; // imutável
let r2 = &s; // imutável — OK, múltiplas são permitidas
// let r3 = &mut s; // ERRO: não pode ter mutable enquanto há imutáveis em uso
println!("{} {}", r1, r2); // r1 e r2 usados aqui, depois disso não são mais acessados
let r3 = &mut s; // OK agora — r1 e r2 não serão mais usados
r3.push_str(", world");
```

### Lifetimes

Lifetimes são anotações que comunicam ao compilador por quanto tempo referências são válidas. Na maioria dos casos são inferidas pelo compilador (lifetime elision rules). Quando não podem ser inferidas, devem ser explícitas:

```rust
fn longest<'a>(x: &'a str, y: &'a str) -> &'a str {
    if x.len() > y.len() { x } else { y }
}
// 'a diz: o retorno vive pelo menos tanto quanto o menor de x e y
```

### Traits

O sistema de traits é o mecanismo de polimorfismo do Rust — análogo a interfaces em Go ou typeclasses em Haskell. Traits definem comportamento compartilhado que tipos podem implementar.

```rust
trait Greet {
    fn hello(&self) -> String;
    fn goodbye(&self) -> String { // implementação padrão
        String::from("Goodbye!")
    }
}

struct Person { name: String }

impl Greet for Person {
    fn hello(&self) -> String {
        format!("Hello, {}!", self.name)
    }
}
```

**Trait objects** (`&dyn Trait`, `Box<dyn Trait>`) permitem polimorfismo em runtime (dynamic dispatch). **Generics com bounds** (`T: Trait`) permitem polimorfismo em compile time (static dispatch, monomorphization).

### Enums e Pattern Matching

Enums em Rust são **algebraic data types** — cada variante pode carregar dados diferentes. Combinados com `match`, eliminam a necessidade de null pointers e exceções:

```rust
enum Result<T, E> {
    Ok(T),
    Err(E),
}

enum Option<T> {
    Some(T),
    None, // sem null — None é explícito no tipo
}

// Pattern matching é exaustivo — o compilador exige que todos os casos sejam tratados
match result {
    Ok(value) => println!("Sucesso: {}", value),
    Err(e) => eprintln!("Erro: {}", e),
}
```

### Macros

Rust tem dois sistemas de macros: **declarative macros** (`macro_rules!`) para substituição de padrões, e **procedural macros** (derive, attribute, function-like) que operam sobre a AST do programa. Macros permitem gerar código em tempo de compilação sem overhead de runtime.

```rust
// Declarative macro
macro_rules! vec {
    ($($x:expr),*) => {
        {
            let mut v = Vec::new();
            $(v.push($x);)*
            v
        }
    };
}

// Procedural macro (derive)
#[derive(Debug, Clone, Serialize, Deserialize)]
struct Config {
    host: String,
    port: u16,
}
```

---

## 3. O Sistema de Edições

Rust usa **Editions** para fazer mudanças incompatíveis de forma controlada e opt-in. Cada crate declara sua edition no `Cargo.toml`. Crates com editions diferentes podem interoperar sem problemas — a edition é uma propriedade do crate, não do compilador.

### Edition 2015 (padrão do Rust 1.0)

A edition original. Toda funcionalidade herdada de 1.0. Ainda suportada e válida.

### Edition 2018 (Rust 1.31, dezembro 2018)

**A maior evolução desde 1.0.** Mudanças principais:

- **Sistema de módulos simplificado** — `use crate::foo` em vez de `use ::foo`. Caminhos relativos tornados explícitos. `extern crate` removido na maioria dos casos.
- **`async`/`await` preparado** como keywords reservadas (implementação viria em 1.39).
- **NLL (Non-Lexical Lifetimes)** habilitado por padrão — borrow checker mais preciso baseado no fluxo de controle real, não em blocos lexicais. Muitos falsos positivos do borrow checker antigo eliminados.
- **`impl Trait`** em posição de argumento — `fn foo(x: impl Display)` em vez de `fn foo<T: Display>(x: T)`.
- **Dyn Trait** obrigatório — `dyn Trait` em vez de apenas `Trait` para trait objects.
- **`?` operator** em mais contextos.

### Edition 2021 (Rust 1.56, outubro 2021)

- **Closures capturam por field** — em vez de capturar a struct inteira, uma closure `move || { println!("{}", x.name) }` captura apenas `x.name`, não `x` inteiro. Resolve muitos conflitos de borrow checker com closures.
- **`IntoIterator` para arrays** — `[1, 2, 3].into_iter()` agora itera sobre valores, não referências.
- **Prelude atualizado** — `TryInto`, `TryFrom`, `FromIterator` entram no prelude automaticamente.
- **Disjoint capture in closures** — já mencionado acima, é a principal mudança.

### Edition 2024 (Rust 1.85, fevereiro 2025)

- **`async fn` em traits** (estabilizado) — `async fn` em trait definitions sem workarounds.
- **`gen` blocks** — geradores experimentais com `yield`.
- **`let chains`** — `if let Some(x) = foo() && x > 0 { }` como expressão única.
- **Prelude atualizado** — `Future` e `IntoFuture` no prelude.
- Várias melhorias de ergonomia em lifetimes e pattern matching.

---

## 4. FFI, Unsafe e Interoperabilidade com C/C++

### O que é `unsafe` em Rust

`unsafe` não desativa as verificações do compilador — ele habilita um conjunto específico de operações adicionais que o compilador não pode verificar automaticamente:

1. Desreferenciar um ponteiro raw (`*const T`, `*mut T`)
2. Chamar funções `unsafe` (incluindo FFI)
3. Acessar ou modificar estado mutável global (`static mut`)
4. Implementar traits `unsafe` (`Send`, `Sync` manualmente)
5. Acessar campos de `union`

O código `unsafe` é uma promessa do programador ao compilador: "eu verifiquei manualmente que isso é seguro, mesmo que você não possa verificar". O objetivo é isolar invariantes não verificáveis em blocos mínimos, não usar `unsafe` amplamente.

```rust
let mut x = 5;
let raw = &mut x as *mut i32; // criar ponteiro raw é seguro

unsafe {
    *raw = 10; // desreferenciar ponteiro raw requer unsafe
}

println!("{}", x); // 10
```

### FFI com C

Rust pode chamar C diretamente sem ferramenta intermediária (equivalente ao CGO no Go, mas sem compilador C separado — o linker cuida disso).

```rust
// Declarar interface C
extern "C" {
    fn strlen(s: *const std::ffi::c_char) -> usize;
    fn malloc(size: usize) -> *mut std::ffi::c_void;
    fn free(ptr: *mut std::ffi::c_void);
}

// Chamar função C
fn main() {
    let s = b"hello\0".as_ptr() as *const std::ffi::c_char;
    let len = unsafe { strlen(s) };
    println!("len = {}", len); // 5
}
```

Para bibliotecas com muitas funções C, o processo manual é impraticável. A ferramenta **`bindgen`** gera bindings Rust automaticamente a partir de headers C:

```bash
# Gera bindings/mod.rs a partir de header.h
bindgen header.h -o bindings.rs
```

### Exportar Rust para C

```rust
// lib.rs
#[no_mangle]  // garante que o nome da função não seja mangled
pub extern "C" fn add(a: i32, b: i32) -> i32 {
    a + b
}

// Cargo.toml
[lib]
crate-type = ["cdylib"]  // gera .so / .dll
// ou "staticlib" para .a
```

Com `cargo build`, produz `libmylib.so` (Linux) ou `mylib.dll` (Windows) — consumível por C, Python (ctypes), Ruby (ffi), Go (CGO), etc.

### Interoperabilidade com C++

C++ é mais complexo que C por causa de name mangling, classes, templates e ABI instável. As abordagens:

**`cxx` crate** (2020–presente) — bridge tipada e segura entre Rust e C++. Você define a interface em um arquivo `.rs` com anotações, e `cxx` gera código bridge para ambos os lados. Muito mais seguro que FFI manual.

```rust
#[cxx::bridge]
mod ffi {
    unsafe extern "C++" {
        include!("mylib.h");
        type MyClass;
        fn process(self: &MyClass, input: &str) -> String;
    }
    extern "Rust" {
        fn rust_callback(data: &[u8]) -> bool;
    }
}
```

**`autocxx`** — geração automática de bindings C++ via `cxx`, similar ao `bindgen` para C puro.

**`cpp` crate** — alternativa que permite inline C++ dentro de Rust (menos usado).

### Diferenças fundamentais entre FFI em Rust vs CGO em Go

| Aspecto                 | Rust FFI                      | Go CGO                          |
| ----------------------- | ----------------------------- | ------------------------------- |
| Compilador C necessário | Não — linker cuida            | Sim — gcc/clang obrigatório     |
| Cross-compilation       | Funciona (com cross-linker)   | Quebra com CGO_ENABLED=1        |
| Overhead de chamada     | ~1ns (sem runtime overhead)   | ~50-3000ns (scheduler overhead) |
| GC interfere            | Não tem GC                    | GC Go não gerencia memória C    |
| Safety                  | `unsafe {}` isola a interface | Nenhuma garantia do compilador  |
| C++                     | Via `cxx` crate               | Frágil, via preamble            |

### `build.rs` — Scripts de Build

Rust permite scripts de build em `build.rs` que rodam antes da compilação. Essencial para compilar código C junto com Rust, gerar bindings, ou configurar flags de linker:

```rust
// build.rs
fn main() {
    // Compilar biblioteca C junto com o crate Rust
    cc::Build::new()
        .file("src/mylib.c")
        .include("include/")
        .flag("-O2")
        .compile("mylib");

    // Dizer ao Rust onde procurar a lib
    println!("cargo:rustc-link-lib=static=mylib");
    println!("cargo:rerun-if-changed=src/mylib.c");
}
```

O crate **`cc`** encapsula a compilação de C/C++/Assembly dentro de build scripts, sendo a forma padrão de incluir código C em crates Rust.

---

## 5. Linha do Tempo: Versões do Rust com Ecossistema Paralelo

### 2006–2010: Origem pessoal e experimentos iniciais

Graydon Hoare trabalha no Rust como projeto pessoal. A linguagem tinha conceitos radicalmente diferentes dos atuais: garbage collector opcional, "typestates" (estados de tipo rastreados pelo compilador), múltiplos tipos de ponteiros especializados, e uma sintaxe muito mais próxima de ML/OCaml. O borrow checker ainda não existia na forma atual — era um sistema diferente chamado "region-based memory management".

**Estado do ecossistema**: Apenas código experimental interno. Sem package manager, sem comunidade.

---

### 2009–2011: Mozilla assume e Servo começa

Mozilla patrocina oficialmente o projeto. O time Servo começa desenvolvimento simultâneo — cada nova feature do Rust era validada pelo que o Servo precisava. Isso moldou profundamente a linguagem: concorrência, performance e segurança de memória se tornaram os pilares porque o Servo precisava deles.

**Rust 0.1 (janeiro 2012)**: Primeiro release público. A linguagem tinha:

- Garbage collector como opção (ao lado do borrow checker)
- Segmented stacks (não contiguous)
- Múltiplos tipos de ponteiros: `@T` (GC pointer), `~T` (owned pointer), `&T` (borrowed)
- Sintaxe muito diferente da atual (`fn main() { }` ainda não era assim)
- Sem `cargo`
- Sem `crates.io`

**Estado do ecossistema 2012**: Praticamente inexistente. Alguns experimentos de libs. Sem gerenciamento de dependências.

---

### 2012–2014: A grande simplificação pré-1.0

Este período foi marcado por **remoções radicais** da linguagem. O time Rust tomou decisões difíceis para simplificar:

- **GC removido** (2013) — a linguagem passou a ser puramente ownership-based. `@T` (GC pointer) eliminado completamente.
- **Typestates removidos** — eram complexos demais na prática.
- **Sintaxe de closures simplificada** — de `|x: int| x * 2` para `|x| x * 2` com inferência de tipo.
- **Sintaxe unificada de ponteiros** — `~T` virou `Box<T>`, `@T` eliminado, apenas `&T` e `&mut T` para referências.
- **`channels` de task** simplificados.

**Rust 0.9 (janeiro 2014)**: Introdução do modelo de `Send + Sync` traits para segurança em concorrência — tipos `Send` podem ser enviados entre threads, `Sync` podem ser acessados de múltiplas threads simultaneamente. Isso formalizou o "fearless concurrency".

**`Cargo` (2014)**: Package manager e build system do Rust. Uma das melhores decisões do projeto — em vez de um ecosistema fragmentado como Python (pip vs conda vs pipenv) ou JavaScript (npm vs yarn vs pnpm), Rust nasceu com uma ferramenta de build universal e oficial.

**`crates.io` (dezembro 2014)**: Registry público de packages. Previamente lançado em beta.

**Estado do ecossistema 2014:**

- `serde` 0.x — serialização/deserialização. Uma das libs mais importantes que existiriam no ecossistema.
- Primeiros experimentos de web frameworks.
- `cargo` muda como código Rust é distribuído para sempre.

---

### Maio 2015: Rust 1.0 — Estabilidade e Compatibilidade

**15 de maio de 2015.** Primeiro release estável. Rust estabelece sua versão de compatibilidade: código que compila em Rust 1.x compilará em Rust 1.y para qualquer y > x (com exceção de soundness fixes).

**O que chegou ao 1.0:**

- Ownership + borrow checker completo e estável
- Lifetimes com anotação explícita quando necessário
- Traits e generics com monomorphization
- Pattern matching exaustivo
- `Result<T, E>` e `Option<T>` como base do error handling
- `Vec<T>`, `HashMap<K,V>`, `String`, `str` na stdlib
- `std::thread` para threads OS
- `std::sync` com `Mutex`, `RwLock`, `Arc`, `Condvar`
- `std::io` com traits `Read`, `Write`, `BufRead`
- Macros declarativas (`macro_rules!`)
- `cargo` e `crates.io`
- Sem `async/await` — concorrência era puramente baseada em threads

**O que NÃO estava no 1.0:**

- Sem `async/await` (viria em 1.39)
- Sem generics em const positions (viria em 1.51+)
- Sem `impl Trait` (viria em 1.26)
- Sem procedural macros estáveis (viriam em 1.15 para derive, 1.30 para attribute)
- Sem NLL (Non-Lexical Lifetimes) — viriam em 1.31+
- Sem `?` operator (viria em 1.13)

**Ecossistema em 1.0:**

- `serde` 0.x — serialização em nightly (stable derive viriam depois)
- `hyper` 0.x — HTTP client/server (em desenvolvimento)
- `iron` — framework web (popular mas baseado em API que mudaria)
- Sem Tokio ainda (viria em 2016)
- `rustfmt` em desenvolvimento
- `clippy` em desenvolvimento

**Como se fazia concorrência antes do async**: threads OS via `std::thread`. Channels via `std::sync::mpsc`. Sem suporte a I/O não-bloqueante de alta performance na stdlib — para isso, usavam `mio` (metal I/O) diretamente. Padrão callback-based ou thread-per-connection.

---

### 2015–2016: Rust 1.1 a 1.14 — Consolidação e Ergonomia

**Rust 1.2 (agosto 2015)**: Melhorias de performance do compilador. Benchmarks começam a mostrar Rust competindo com C/C++ em CPU-bound tasks.

**Rust 1.4 (outubro 2015)**: `rustfmt` torna-se ferramenta oficial. Padronização de estilo — sem guerras de formatação.

**Rust 1.5 (dezembro 2015)**: Melhorias em `std::fs` e paths. Atributos em loops (`'label: loop { break 'label; }`).

**Rust 1.6 (janeiro 2016)**: `drain()` em coleções. Primeiras discussões públicas sobre `async`.

**Rust 1.7 (março 2016)**: Melhorias em stdlib. `HashMap` com Siphash por padrão.

**Rust 1.9 (maio 2016)**: `std::panic::catch_unwind` — capturar panics. Melhorias de mensagens de erro.

**Rust 1.10 (julho 2016)**: `cdylib` crate type — bibliotecas C-ABI em Rust. Essencial para WASM e plugins.

**Rust 1.12 (setembro 2016)**: **Mensagens de erro completamente redesenhadas.** O compilador Rust era famoso por mensagens de erro incompreensíveis. 1.12 iniciou uma era de erros amigáveis com spans coloridos, sugestões automáticas e `error[E0XXX]: rustc --explain E0XXX`. Isso mudou fundamentalmente a experiência de aprendizado.

**Rust 1.13 (novembro 2016)**: **`?` operator** — `try!()` macro substituído pelo operador `?` para propagação de erros. Código async-like com `Result` ficou muito mais legível.

```rust
// Antes de 1.13
fn read_file(path: &str) -> Result<String, io::Error> {
    let mut f = try!(File::open(path));
    let mut s = String::new();
    try!(f.read_to_string(&mut s));
    Ok(s)
}

// Depois de 1.13
fn read_file(path: &str) -> Result<String, io::Error> {
    let mut f = File::open(path)?;
    let mut s = String::new();
    f.read_to_string(&mut s)?;
    Ok(s)
}
```

**Rust 1.14 (dezembro 2016)**: **Rustup 1.0** — toolchain manager oficial. Gerencia versões de Rust (stable/beta/nightly), targets para cross-compilation, e componentes (rustfmt, clippy, rust-analyzer). Resolvia o problema de instalar e gerenciar múltiplas versões de Rust.

**Ecossistema 2015–2016:**

- **`serde` 1.0 em preparação** — framework de serialização mais importante do ecossistema
- **`tokio` 0.1 (2016)** — runtime assíncrono baseado em `mio`. Primeira versão. API bastante diferente da atual.
- **`hyper` 0.9** — HTTP baseado em futures
- **`actix-web` primeiras versões (2017)** — framework baseado em actor model
- **`diesel` 0.x (2016)** — ORM typesafe em Rust
- **`rayon` 0.x (2016)** — data parallelism library

---

### 2017: Rust 1.15–1.22 — Macros e Derive

**Rust 1.15 (fevereiro 2017)**: **Procedural macros (derive)** estabilizados em stable. Este foi o desbloqueio do `serde`. Antes, `#[derive(Serialize, Deserialize)]` requeria nightly. Com 1.15, passou a funcionar em stable — e o ecossistema explodiu.

```rust
// Antes de 1.15: apenas em nightly
// Depois de 1.15: funciona em stable!
#[derive(Debug, Serialize, Deserialize)]
struct Config {
    host: String,
    port: u16,
    max_connections: Option<u32>,
}
```

**Rust 1.16 (março 2017)**: `cargo check` — verifica o código sem compilar. 5-10x mais rápido que `cargo build` para verificar erros durante desenvolvimento.

**Rust 1.17 (abril 2017)**: Build system movido de `make` para `cargo`. O compilador Rust passou a ser compilado com cargo — "dogfooding" importante.

**Rust 1.20 (agosto 2017)**: **Métodos e constantes associadas a traits** — `const` em trait definitions.

**Rust 1.21 (outubro 2017)**: `eprint!` e `eprintln!` para stderr. Melhorias de performance do compilador (~20% mais rápido).

**Ecossistema 2017:**

- **`serde` 1.0 estável (maio 2017)** — milestone fundamental. Serialize/Deserialize para qualquer formato (JSON, TOML, YAML, MessagePack, Bincode...) com derive macros. Sem serde, o ecossistema Rust seria radicalmente diferente.
- **`tokio` 0.1 estável** — async I/O via futures 0.1
- **`rayon` 1.0 (2018)** — data parallelism trivial com `par_iter()`
- **`clap` 2.x** — CLI argument parser
- **`log` + `env_logger`** — logging
- **`reqwest` 0.x** — HTTP client de alto nível

---

### Dezembro 2018: Rust 1.31 — Edition 2018

**Esta é a versão mais importante depois de 1.0.**

**Rust 1.31 (dezembro 2018)**: **Edition 2018**. Veja seção 3 para detalhes completos da edition. Além da edition:

- **`const fn`** — funções executáveis em compile time
- **`match` com guards melhorados**
- Tool attributes (`#[rustfmt::skip]`, `#[allow(clippy::...)]`)
- **`extern crate` removido** para a maioria dos usos (edition 2018)

**Ecossistema 1.31:**

- `async-std` em desenvolvimento (alternativa ao Tokio)
- `warp` 0.1 — framework web funcional/composable
- `rocket` 0.3 — framework web ergonômico (ainda em nightly)

---

### 2019: Rust 1.32–1.39 — Async/Await

**Rust 1.32 (janeiro 2019)**: `dbg!()` macro — debug print com localização do arquivo e linha. `std::mem::ManuallyDrop` estabilizado.

**Rust 1.34 (abril 2019)**: `TryFrom`/`TryInto` estabilizados — conversões que podem falhar com `Result`. Caminhos `..` em pattern matching expandidos.

**Rust 1.36 (julho 2019)**: **`Future` trait estabilizado na stdlib.** Não é `async/await` ainda — é a base. O `Future` define o contrato para computações assíncronas. Habilitou que crates preparassem suas APIs para `async/await`.

```rust
// Future trait (estabilizado em 1.36)
pub trait Future {
    type Output;
    fn poll(self: Pin<&mut Self>, cx: &mut Context<'_>) -> Poll<Self::Output>;
}
```

**Rust 1.37 (agosto 2019)**: `#[must_use]` em funções. `cargo vendor` para dependências offline. Referências para arrays em pattern matching.

**Rust 1.39 (novembro 2019)**: **`async`/`await` ESTABILIZADOS.** Um dos maiores milestones da história do Rust.

```rust
use tokio::time::{sleep, Duration};

// async fn retorna um Future implicitamente
async fn fetch_data(url: &str) -> Result<String, reqwest::Error> {
    let response = reqwest::get(url).await?; // .await "pausa" até o Future completar
    response.text().await
}

#[tokio::main] // macro que inicializa o runtime Tokio
async fn main() {
    match fetch_data("https://example.com").await {
        Ok(body) => println!("Body: {}", body),
        Err(e) => eprintln!("Error: {}", e),
    }
}
```

O `async/await` do Rust é fundamentalmente diferente do de outras linguagens. Não existe runtime embutido na linguagem — você escolhe o executor (`tokio`, `async-std`, `smol`). Isso mantém o runtime de IA/cloud fora do binário padrão e permite escolher o executor mais adequado (incluindo executores customizados para embedded).

**Ecossistema 1.39:**

- **`tokio` 0.2 (novembro 2019)** — reescrito para async/await. API completamente diferente de 0.1.
- **`async-std` 1.0 (setembro 2019)** — alternativa ao Tokio com API espelhando a stdlib mas async
- **`reqwest` 0.10 (dezembro 2019)** — HTTP client async/await
- **`sqlx` 0.1 (outubro 2019)** — SQL queries async com compile-time checking

---

### 2020: Rust 1.40–1.49 — Consolidação Async

**Rust 1.40 (dezembro 2019/jan 2020)**: `todo!()` macro. `#[non_exhaustive]` — marca enums/structs como "pode ter mais variantes no futuro". Essencial para APIs públicas que evoluem.

**Rust 1.42 (março 2020)**: **Subslice patterns** — `[first, .., last]` em pattern matching. `matches!()` macro.

**Rust 1.43 (abril 2020)**: `float::INFINITY`, `float::NEG_INFINITY`, `float::NAN` como consts associadas.

**Rust 1.44 (junho 2020)**: `cargo tree` integrado — visualiza árvore de dependências. Async `TcpListener` na stdlib.

**Rust 1.45 (julho 2020)**: Fixes críticos em casting (`as`) de floats para inteiros — antes podiam causar UB, agora são saturating.

**Rust 1.46 (agosto 2020)**: `const fn` muito expandido — `if`, `match`, loops em `const fn`. Mais código executável em compile time.

**Rust 1.47 (outubro 2020)**: **`TraitObject` sem dyn removido** (edition 2021 prep). Suporte completo a LLVM 11.

**Rust 1.49 (dezembro 2020)**: **Tier 1 para aarch64** (ARM64 Linux). Apple Silicon support começando. ARM64 Linux se torna plataforma de primeira classe.

**Ecossistema 2020:**

- **`tokio` 0.3 (outubro 2020)** — iteração de API
- **`axum` em desenvolvimento** (seria lançado em 2021)
- **`tauri` 0.x (2020)** — desktop app com WebView
- **`egui` 0.1 (dezembro 2020)** — immediate mode GUI em Rust puro
- **`sled` (banco embutido)** ganha maturidade
- **`tracing` 0.1** — structured logging e instrumentação

---

### 2021: Rust 1.50–1.56 — Edition 2021 e Maturidade

**Rust 1.51 (março 2021)**: **Const generics MVP** — tipos e funções parametrizados por constantes inteiras. Desbloqueou todo um universo de APIs type-safe:

```rust
// Array de tamanho N conhecido em compile time
fn first_n<T, const N: usize>(arr: [T; N]) -> Option<&T> {
    if N > 0 { Some(&arr[0]) } else { None }
}

// Qualquer tamanho, verificado em compile time
let a: [i32; 3] = [1, 2, 3];
let b: [i32; 10] = [0; 10];
```

**Rust 1.52 (maio 2021)**: Melhorias incrementais de compilação. `String::drain_filter` experimental.

**Rust 1.53 (junho 2021)**: **`IntoIterator` para arrays** — `for x in [1, 2, 3]` itera por valor (edition 2021 prep). Pattern matching em `|` (or patterns) em todas as posições.

**Rust 1.54 (julho 2021)**: Macros em attribute arguments. `BTreeMap::retain`.

**Rust 1.56 (outubro 2021)**: **Edition 2021**. Veja seção 3. Além da edition: `Cargo.lock` em gitignore por padrão para libraries. `std::thread::available_parallelism()`.

**Fevereiro 2021**: **Rust Foundation criada.** Mozilla transfere a governança do Rust para uma fundação independente com membros fundadores: AWS, Huawei, Google, Microsoft, Mozilla. O Rust deixa de ser "a linguagem do Mozilla" e se torna uma fundação neutra.

**Ecossistema 2021:**

- **`tokio` 1.0 (dezembro 2020/início 2021)** — API estável. Marco histórico.
- **`axum` 0.1 (julho 2021)** — framework web do time Tokio. `tower::Service` como base.
- **`tauri` 1.0 RC** — apps desktop com WebView nativo
- **`egui` 0.13+** — immediate mode GUI crescendo rapidamente
- **`sqlx` 0.5** — queries async com compile-time checking maduro
- **`iced` 0.3** — GUI inspirado em Elm
- **`bevy` 0.5** — game engine em Rust

---

### 2022: Rust 1.57–1.65 — Generic Associated Types e let-else

**Rust 1.58 (janeiro 2022)**: **Captured identifiers em format strings** — `let name = "world"; println!("Hello, {name}!")` em vez de `println!("Hello, {}", name)`.

**Rust 1.60 (abril 2022)**: **Cargo features com `dep:` prefix** — controle mais preciso de features condicionais. `--timings` flag para profiling de compilação.

**Rust 1.62 (junho 2022)**: `cargo add` integrado — adiciona dependências via CLI sem editar `Cargo.toml` manualmente.

**Rust 1.63 (agosto 2022)**: **Scoped threads** — `std::thread::scope()` permite spawnar threads que referenciam dados locais com lifetime garantido pelo compilador. Resolve o problema de threads 'static.

```rust
let data = vec![1, 2, 3];
std::thread::scope(|s| {
    s.spawn(|| {
        println!("Thread: {:?}", &data); // pode usar &data sem Arc!
    });
}); // todas as threads do scope terminam aqui
// data ainda é válido aqui
```

**Rust 1.65 (novembro 2022)**: **Generic Associated Types (GATs)** estabilizados. Feature que levou anos de desenvolvimento, GATs permitem tipos associados em traits serem parametrizados:

```rust
trait Container {
    type Item<'a> where Self: 'a; // GAT — Item tem lifetime parameter
    fn get<'a>(&'a self, idx: usize) -> Option<Self::Item<'a>>;
}
```

**`let-else` statement** (1.65):

```rust
// Sem let-else
let value = match some_option {
    Some(v) => v,
    None => return, // early return no else
};

// Com let-else (1.65)
let Some(value) = some_option else { return };
// value disponível aqui com o tipo unwrapped
```

**Ecossistema 2022:**

- **`tauri` 1.0 estável (junho 2022)** — marco histórico para desktop Rust
- **`axum` 0.5** com extractor API refinada
- **`pyo3` 0.16+** — bindings Rust para Python. Rust como extensão Python
- **`rustls` 0.20** — TLS em Rust puro (sem OpenSSL)
- **`bevy` 0.8** — game engine crescendo rapidamente
- **`slint` 0.2 (antigo SixtyFPS)** — GUI para embedded e desktop

---

### 2023: Rust 1.66–1.74 — Estabilizações e Ergonomia

**Rust 1.66 (dezembro 2022)**: **Discriminants de enum explícitos com fields** — `enum Foo { A(u32) = 1, B(u32) = 2 }`. `black_box` estabilizado para benchmarks.

**Rust 1.67 (janeiro 2023)**: `Duration::checked_add_duration`, melhorias async internals.

**Rust 1.70 (junho 2023)**: **`OnceLock`** — lazy initialization thread-safe. **`IsTerminal`** trait para detectar se stdout/stderr é terminal. Cargo default para `sparse` registry protocol — downloads de metadata muito mais rápidos.

**Rust 1.71 (julho 2023)**: **C-unwind ABI** estabilizado — permite unwinding de panics através de boundaries FFI de forma segura.

**Rust 1.73 (outubro 2023)**: `thread::scope` melhorado. Mensagens de panic incluem o thread name.

**Rust 1.74 (novembro 2023)**: **`lint` configuration via Cargo.toml** — `[lints]` section permite configurar lints por workspace. `Cargo.toml` workspace lints propagam para todos os membros.

**Ecossistema 2023:**

- **`axum` 0.7 (novembro 2023)** — usa `hyper` 1.0, breaking changes importantes
- **`tokio` 1.35** — stable e amplamente adotado
- **`dioxus` 0.4** — framework UI React-like para desktop/web
- **`leptos` 0.5** — framework SSR/CSR para web com signals
- **`tauri` 2.0 beta** — mobile support (iOS/Android)
- **`async-graphql` 6.x** — GraphQL server em Rust async
- **`burn`** — deep learning framework em Rust

---

### 2024: Rust 1.75–1.82 — Async fn in Traits e Return Position impl Trait

**Rust 1.75 (dezembro 2023)**: **`async fn` e `return position impl Trait` em traits** estabilizados — sem workarounds, sem `Pin<Box<dyn Future>>`:

```rust
// Antes de 1.75
trait Fetcher {
    fn fetch(&self, url: &str) -> Pin<Box<dyn Future<Output = String> + '_>>;
}

// Depois de 1.75
trait Fetcher {
    async fn fetch(&self, url: &str) -> String;
    // ou
    fn fetch(&self, url: &str) -> impl Future<Output = String> + '_;
}
```

**Rust 1.77 (março 2024)**: **`C-string literals`** — `c"hello"` como sintaxe nativa para strings C null-terminated. `offset_of!` macro estabilizado.

**Rust 1.78 (maio 2024)**: **`#[diagnostic]` attribute namespace** — permite crates customizarem mensagens de erro do compilador. `unsafe_op_in_unsafe_fn` promovido para warn-by-default.

**Rust 1.79 (junho 2024)**: **Inline `const` expressions** — `let x = const { /* compute at compile time */ };`. Associated type bounds inline.

**Rust 1.80 (julho 2024)**: **`LazyCell` e `LazyLock`** — lazy initialization na stdlib sem dependência de `once_cell`. `exclusive_range_pattern` estabilizado.

**Rust 1.82 (outubro 2024)**: **`&raw const` e `&raw mut`** — ponteiros raw sem criar referência intermediária. Essencial para código FFI seguro com campos unaligned. `impl Trait` em mais posições.

**Ecossistema 2024:**

- **`tauri` 2.0 estável** — iOS e Android suportados
- **`axum` 0.8 (janeiro 2026)** — API estabilizando
- **`rustls` 0.23** — TLS puro Rust amplamente adotado em substituição ao OpenSSL
- **Rust no Linux Kernel** — Rust como segunda linguagem oficial do kernel Linux desde 6.1 (2022), expandindo em 2024
- **Rust no Windows Kernel** — Microsoft usando Rust em componentes do kernel Windows

---

### Fevereiro 2025: Rust 1.85 — Edition 2024

**Rust 1.85**: **Edition 2024** (veja seção 3). Além da edition:

- `async fn` em traits completamente ergonômico
- `gen` blocks (geradores com `yield`) em experimental
- `let chains` em `if let` e `while let`

**Versões 1.86–1.94 (2025)**: Continuação de estabilizações incrementais, melhorias do compilador, novos targets.

---

## 6. Runtime Assíncrono: Tokio e o Ecossistema Async

### O problema que async/await resolve

I/O-bound workloads (web servers, databases, network) passam a maior parte do tempo esperando. Thread-per-connection escala mal — cada thread OS custa ~1MB de stack, e context switches custam ~microsegundos. Async permite **M:N multiplexing**: milhares de tasks pendentes em poucas threads OS.

### Como async funciona em Rust

`async fn` é açúcar sintático para uma função que retorna um tipo que implementa `Future`. O compilador transforma o corpo da função em uma **state machine** que pode ser pausada e resumida pelo executor:

```rust
async fn fetch(url: &str) -> String {
    // compilado para uma state machine com estados:
    // State0: chamou reqwest::get, aguardando
    // State1: get retornou, chamou response.text(), aguardando
    // State2: text retornou, done
    reqwest::get(url).await.unwrap().text().await.unwrap()
}
```

O executor (Tokio, async-std, smol) roda as state machines — ele chama `Future::poll()` em cada task. Quando uma task retorna `Poll::Pending` (ainda esperando I/O), o executor roda outra task. Quando o I/O fica pronto (via epoll/kqueue/IOCP), o executor é notificado e chama `poll()` novamente.

### `tokio` (2016–presente)

O runtime assíncrono mais usado em Rust. Criado por Carl Lerche inicialmente como `mio` (metal I/O — wrapper thin sobre epoll/kqueue/IOCP), depois evoluiu para Tokio com futures e depois async/await.

**Versões importantes:**

- 0.1 (2016) — baseado em futures 0.1, API old-style
- 0.2 (novembro 2019) — reescrito para async/await. API completamente diferente
- 0.3 (outubro 2020) — iteração de estabilização
- **1.0 (dezembro 2020) — API estável**. Marco histórico.

**Componentes do ecossistema Tokio:**

- `tokio` — runtime, scheduler, timers, tasks
- `tokio::net` — TCP, UDP, Unix sockets async
- `tokio::fs` — filesystem async
- `tokio::sync` — primitivos sync async: `Mutex`, `RwLock`, `Semaphore`, `watch`, `broadcast`, `mpsc`, `oneshot`
- `tokio-util` — codecs, framing, I/O utilities
- `tower` — abstrações de middleware (`Service`, `Layer`)
- `hyper` — HTTP/1.1 e HTTP/2 de baixo nível (base do Axum e Reqwest)
- `tonic` — gRPC sobre Tokio/Hyper

### `async-std` (2019–2020s)

Alternativa ao Tokio com API que espelha a stdlib. Em vez de `tokio::fs::read_to_string`, é `async_std::fs::read_to_string`. Menos adotado que Tokio hoje, mas importante como alternativa.

### `smol` (2020–presente)

Runtime async minimalista. Codebase pequena (~1500 linhas), sem macros. Para casos onde o tamanho do binário importa (embedded, WASM).

### `Pin` e por que ele existe

`Pin<P>` é uma garantia de que o valor apontado por `P` não será movido na memória enquanto o `Pin` existe. Isso é necessário porque state machines geradas pelo compilador para `async fn` podem conter auto-referências — se movidas, os ponteiros internos seriam inválidos.

```rust
// Interno ao compilador — você raramente manipula Pin diretamente em código de aplicação
use std::pin::Pin;
use std::future::Future;

fn spawn<F: Future + Send + 'static>(future: F) {
    // Box::pin move o future para a heap e pinna ele
    tokio::spawn(Box::pin(future));
}
```

---

## 7. Frameworks Web

### 7.1 Antes dos frameworks: `hyper` direto (2014–presente)

`hyper` é a biblioteca HTTP de baixo nível que a maioria dos frameworks usa internamente. Usa `tokio` para I/O. Funcional direto mas verboso para APIs REST. Ainda usado em projetos que precisam de controle total ou estão construindo frameworks.

### 7.2 `iron` (2015–2018, abandonado)

Primeiro framework web popular do Rust. Inspirado em Node.js/Connect. Middleware chain com `Handler` traits. Abandonado após async/await tornar a abordagem obsoleta e seus mantenedores moverem para outras prioridades.

**Antes do Iron**: middleware web era feito manualmente sobre `hyper` 0.x com futures 0.1 — extremamente verboso, sem ergonomia, difícil de compor.

### 7.3 `rocket` (2016–presente)

Framework inspirado em Rails/Django. Usa macros de procedimento para routing declarativo e ergonômico. Famoso por requerer nightly Rust por anos (por usar features instáveis) — migrou para stable apenas com a versão 0.5.

```rust
#[macro_use] extern crate rocket;

#[get("/users/<id>")]
fn get_user(id: u64) -> String {
    format!("User {}", id)
}

#[post("/users", data = "<user>")]
fn create_user(user: Json<NewUser>) -> Created<Json<User>> {
    // ...
}

#[launch]
fn rocket() -> _ {
    rocket::build().mount("/", routes![get_user, create_user])
}
```

Rocket 0.5 (2022) finalmente rodou em stable Rust com suporte a async/await. Inclui: form handling, templating, state management, testing utilities. Mais "batteries included" que Axum.

### 7.4 `actix-web` (2017–presente)

O framework web mais rápido do Rust — consistentemente no topo de benchmarks TechEmpower. Baseado no actor model do `actix` (Erlang/Akka-inspired). O criador original abandonou o projeto em 2020 após controvérsia sobre uso de `unsafe` (código unsafe excessivo nos internals). A comunidade assumiu e reescreveu partes.

```rust
use actix_web::{web, App, HttpServer, HttpResponse};

async fn get_user(path: web::Path<u64>) -> HttpResponse {
    HttpResponse::Ok().json(format!("User {}", path.into_inner()))
}

#[actix_web::main]
async fn main() -> std::io::Result<()> {
    HttpServer::new(|| {
        App::new()
            .route("/users/{id}", web::get().to(get_user))
    })
    .bind("127.0.0.1:8080")?
    .run()
    .await
}
```

10-15% mais rápido que Axum em benchmarks sintéticos. Maior ecossistema de middleware. Curva de aprendizado mais íngreme — o actor model adiciona complexidade conceitual.

### 7.5 `warp` (2019–presente)

Framework funcional/composable. Routing via **Filters** que se combinam como funções:

```rust
use warp::Filter;

let hello = warp::path("hello")
    .and(warp::path::param::<String>())
    .and(warp::header("user-agent"))
    .map(|name: String, agent: String| {
        format!("Hello {} from {}", name, agent)
    });

warp::serve(hello).run(([127, 0, 0, 1], 3030)).await;
```

Tipo-seguro mas com curva de aprendizado alta devido aos tipos genéricos complexos dos Filters. Menor que Axum em adoção hoje mas influente no design.

### 7.6 `axum` (2021–presente)

O framework preferido pela maioria dos novos projetos Rust. Criado pelo time do Tokio. Zero macros — usa o sistema de tipos Rust para routing e extractors. Integra com o ecossistema `tower` para middleware.

```rust
use axum::{Router, routing::{get, post}, extract::{Path, State, Json}, http::StatusCode};
use std::sync::Arc;

#[derive(Clone)]
struct AppState {
    db: PgPool,
}

async fn get_user(
    State(state): State<Arc<AppState>>,
    Path(id): Path<u64>,
) -> Result<Json<User>, StatusCode> {
    sqlx::query_as!(User, "SELECT * FROM users WHERE id = $1", id as i64)
        .fetch_one(&state.db)
        .await
        .map(Json)
        .map_err(|_| StatusCode::NOT_FOUND)
}

#[tokio::main]
async fn main() {
    let state = Arc::new(AppState { db: connect().await });
    let app = Router::new()
        .route("/users/:id", get(get_user))
        .with_state(state)
        .layer(TraceLayer::new_for_http()); // tower middleware

    axum::serve(
        tokio::net::TcpListener::bind("0.0.0.0:3000").await.unwrap(),
        app
    ).await.unwrap();
}
```

**Extractors**: qualquer tipo que implementa `FromRequest` pode ser um parâmetro de handler — `Path<T>`, `Query<T>`, `Json<T>`, `State<T>`, `Headers`, `Extension<T>`. Você cria extractors customizados para autenticação, rate limiting, etc.

### 7.7 Frameworks full-stack

**`Leptos`** (2022–presente) — framework reativo full-stack com signals. Compile para WASM no browser ou para SSR no server. Usa macros RSX (similar a JSX) para templates em Rust. Sem JavaScript obrigatório.

**`Dioxus`** (2021–presente) — framework React-like para web, desktop (via Tauri), mobile e TUI. Componentes, hooks, state management em Rust.

**`Loco`** (2023–presente) — framework estilo Rails para Rust. Inclui: ORM (sea-orm), mailers, background jobs, scaffolding. Para quem quer "Rails em Rust".

### 7.8 gRPC

**`tonic`** (2019–presente) — implementação gRPC para Rust sobre Tokio/Hyper. Gera stubs Rust a partir de `.proto` files via `prost`. Suporta streaming bidirecional.

```rust
// Servidor gRPC
#[tonic::async_trait]
impl Greeter for MyGreeter {
    async fn say_hello(&self, req: Request<HelloRequest>) -> Result<Response<HelloReply>, Status> {
        Ok(Response::new(HelloReply {
            message: format!("Hello, {}!", req.into_inner().name),
        }))
    }
}
```

---

## 8. CLI e TUI

### 8.1 `clap` (2015–presente)

O parser de argumentos CLI mais usado em Rust. Derivativo (macro-based) ou builder API.

```rust
use clap::Parser;

#[derive(Parser)]
#[command(version, about)]
struct Cli {
    #[arg(short, long)]
    port: u16,

    #[arg(short, long, default_value = "false")]
    verbose: bool,

    #[command(subcommand)]
    command: Commands,
}

#[derive(Subcommand)]
enum Commands {
    Serve { #[arg(short)] host: String },
    Migrate,
}

fn main() {
    let cli = Cli::parse();
    match cli.command {
        Commands::Serve { host } => serve(host, cli.port),
        Commands::Migrate => migrate(),
    }
}
```

v2 (2015-2021): builder API dominante. v3 (2022): derive API adicionada. v4 (2022): refactor, derive como API primária.

**Antes do Clap**: parsing manual de `std::env::args()` ou `getopts` (port do utilitário Unix). Verbose, sem ajuda automática, sem subcomandos.

### 8.2 `ratatui` (2023–presente, fork de `tui-rs`)

Framework TUI (Text User Interface) baseado no modelo retained-mode. Widgets, layouts, event handling. Fork do `tui-rs` abandonado.

```rust
use ratatui::{widgets::{Block, Borders, Paragraph}, layout::{Layout, Constraint}};

fn render(frame: &mut Frame) {
    let layout = Layout::vertical([Constraint::Length(3), Constraint::Min(0)]);
    let [header, body] = layout.areas(frame.area());

    frame.render_widget(
        Paragraph::new("Vectora v1.0").block(Block::bordered().title("Status")),
        header,
    );
    frame.render_widget(Block::bordered().title("Chat"), body);
}
```

**Antes do ratatui/tui-rs**: TUIs em Rust usavam `crossterm` ou `termion` diretamente (escape codes ANSI) — baixo nível, sem layout engine, muito código boilerplate.

### 8.3 `crossterm` (2018–presente)

Biblioteca de terminal cross-platform. Manipulação de cursor, cores, estilos, eventos de teclado/mouse. Base para ratatui e outras TUI libs. Funciona no Windows (sem ANSI nativo antigo) e Unix.

### 8.4 `indicatif` (2017–presente)

Progress bars e spinners para CLI. Simples e polido:

```rust
use indicatif::{ProgressBar, ProgressStyle};

let pb = ProgressBar::new(100);
pb.set_style(ProgressStyle::default_bar()
    .template("{spinner} [{bar:40}] {pos}/{len} {msg}")?);
for i in 0..100 {
    pb.set_position(i);
    pb.set_message(format!("Processing file {}", i));
}
pb.finish_with_message("Done!");
```

### 8.5 `dialoguer` (2018–presente)

Prompts interativos para CLI: confirmação, seleção de lista, input de texto, senha, checkboxes. Usa `console` internamente.

---

## 9. GUI e Desktop

### 9.1 Antes dos frameworks Rust: Bindings para GTK/Qt (2015–2018)

**`gtk-rs`** — bindings Rust para GTK3/GTK4 via FFI. Unsafe internamente mas com API safe exposta. Requer libgtk instalada no sistema. Funciona mas não é idiomático Rust — a GObject system do GTK com signals e callbacks não mapeia naturalmente para ownership.

**`qt-rs`/`ritual`** — bindings para Qt. Extremamente complexos, nunca se tornaram mainstream em Rust.

**`winit`** (2017–presente) — criação de janelas cross-platform sem GUI toolkit completo. Base para Bevy, iced, egui. Cuida de eventos de janela (resize, input, focus) mas não de widgets.

**Padrão 2015-2019**: Rust para backend + frontend em Electron/Qt/GTK em outra linguagem. Dois binários comunicando via HTTP local ou IPC. Similar ao que Go fazia.

### 9.2 `egui` (2020–presente)

**Immediate Mode GUI** — widgets são recriados a cada frame, sem estado persistente de UI. Cada frame, você descreve o que renderizar e a lib cuida do diff e rendering.

```rust
fn update(&mut self, ctx: &egui::Context, _frame: &mut eframe::Frame) {
    egui::CentralPanel::default().show(ctx, |ui| {
        ui.heading("Vectora");
        ui.horizontal(|ui| {
            ui.label("Query:");
            ui.text_edit_singleline(&mut self.query);
        });
        if ui.button("Search").clicked() {
            self.results = self.db.search(&self.query);
        }
        for result in &self.results {
            ui.label(result);
        }
    });
}
```

Renderização via OpenGL/WebGPU/WASM. **Sem dependências de sistema** — bundle 100% standalone. Ideal para ferramentas de desenvolvedor, debug UIs, aplicações simples. Não tem look nativo do SO — aparência própria do egui.

### 9.3 `iced` (2020–presente)

Framework GUI baseado no **Elm Architecture** (Model-Update-View). Retained mode. Renderização via WGPU (WebGPU). Suporta desktop e WASM.

```rust
struct Counter { value: i32 }

#[derive(Debug, Clone)]
enum Message { Increment, Decrement }

impl Counter {
    fn update(&mut self, message: Message) {
        match message {
            Message::Increment => self.value += 1,
            Message::Decrement => self.value -= 1,
        }
    }

    fn view(&self) -> Element<Message> {
        column![
            button("Increment").on_press(Message::Increment),
            text(self.value),
            button("Decrement").on_press(Message::Decrement),
        ].into()
    }
}
```

Arquitetura mais limpa que egui para apps complexos. Look próprio (não nativo do SO). Acessibilidade ainda limitada.

### 9.4 `tauri` (2020–presente)

Equivalente Rust do Wails para Go — frontend web + backend Rust, usando WebView nativo do SO. Sem Chromium embutido.

```rust
// Rust backend
#[tauri::command]
async fn search_documents(query: String, state: State<'_, AppState>) -> Result<Vec<Document>, String> {
    state.db.search(&query).await.map_err(|e| e.to_string())
}

fn main() {
    tauri::Builder::default()
        .manage(AppState::new())
        .invoke_handler(tauri::generate_handler![search_documents])
        .run(tauri::generate_context!())
        .expect("error while running tauri application");
}
```

```typescript
// Frontend TypeScript — Tauri IPC
import { invoke } from "@tauri-apps/api/core";

const results = await invoke<Document[]>("search_documents", {
  query: "godot signals",
});
```

**Versões importantes:**

- 1.0 (junho 2022) — estável para desktop
- 2.0 (outubro 2024) — mobile (iOS, Android), nova IPC API, plugins

**Diferença Tauri vs Wails**: Tauri usa JSON RPC via postMessage do WebView. Wails usa bindings JS diretos gerados automaticamente a partir do Go. Tauri tem ecossistema de plugins muito mais rico e suporte mobile. Wails é mais simples de configurar.

### 9.5 `slint` (2021–presente, antes SixtyFPS)

Toolkit GUI com **linguagem declarativa própria** (`.slint` files) para definir UI. Backend Rust para lógica. Foco em: embedded systems, microcontroladores, desktop. Motor de renderização proprio em OpenGL/DirectX.

```slint
// main.slint
export component MainWindow inherits Window {
    title: "Vectora";
    in-out property <string> query;
    callback search(string);

    VerticalBox {
        LineEdit { text <=> query; }
        Button { text: "Search"; clicked => { search(query); } }
    }
}
```

```rust
// main.rs
fn main() {
    let ui = MainWindow::new().unwrap();
    ui.on_search(|query| println!("Searching: {}", query));
    ui.run().unwrap();
}
```

Licença: GPLv3 (gratuito para apps open source) ou comercial. Muito utilizado em produtos embedded comerciais.

### 9.6 `dioxus` (2021–presente)

Framework React-like para web, desktop, mobile e TUI — tudo da mesma codebase:

```rust
fn App() -> Element {
    let mut count = use_signal(|| 0);
    rsx! {
        h1 { "Count: {count}" }
        button { onclick: move |_| count += 1, "Increment" }
        button { onclick: move |_| count -= 1, "Decrement" }
    }
}
```

Para desktop usa Tauri internamente. Para web compila para WASM. Para TUI usa ratatui. Para mobile ainda experimental.

---

## 10. Banco de Dados e Persistência

### 10.1 `diesel` (2016–presente)

ORM e query builder **compile-time safe** — queries SQL são verificadas contra o schema em tempo de compilação. Sem erros de tipagem em runtime.

```rust
use diesel::prelude::*;

// Schema inferido do banco via diesel print-schema
table! {
    users (id) {
        id -> Integer,
        name -> Text,
        email -> Text,
    }
}

// Query com verificação em compile time
let results = users::table
    .filter(users::name.like("%Bruno%"))
    .select(User::as_select())
    .load(&mut conn)?;
```

Não é async nativamente (v1/v2). `diesel-async` adiciona suporte async mas com limitações. Para workloads async-first, `sqlx` é preferível.

### 10.2 `sqlx` (2019–presente)

SQL queries com **verificação em compile time contra banco real**. Async-first. Suporta PostgreSQL, MySQL, SQLite, MSSQL.

```rust
use sqlx::PgPool;

// Query verificada contra o banco em COMPILE TIME via sqlx offline mode
let user = sqlx::query_as!(
    User,
    "SELECT id, name, email FROM users WHERE id = $1",
    user_id
)
.fetch_one(&pool)
.await?;

// Ou com macros (requer banco rodando durante compilação)
let users: Vec<User> = sqlx::query_as!(User, "SELECT * FROM users")
    .fetch_all(&pool)
    .await?;
```

**`sqlx offline mode`**: gera arquivo JSON com os tipos das queries em tempo de dev. Em CI/produção, verifica sem precisar de banco conectado.

### 10.3 `sea-orm` (2021–presente)

ORM async construído sobre `sqlx`. API mais alta que `sqlx` puro. Suportado pelo Loco framework.

### 10.4 `rusqlite` (2014–presente)

Bindings para SQLite via FFI. Síncrono. O equivalente Rust do `mattn/go-sqlite3`.

```rust
use rusqlite::{Connection, Result};

let conn = Connection::open("vectora.db")?;
conn.execute("CREATE TABLE IF NOT EXISTS workspaces (id INTEGER PRIMARY KEY, name TEXT)", [])?;
conn.execute("INSERT INTO workspaces (name) VALUES (?1)", ["Godot 4.2"])?;
```

### 10.5 `sled` (2018–presente)

Banco key-value embarcado em Rust puro. Similar ao BoltDB do Go. B-tree otimizado para flash/SSD. API async-friendly.

```rust
let db = sled::open("my_db")?;
db.insert("key", "value")?;
let val = db.get("key")?.unwrap();
```

### 10.6 `redb` (2022–presente)

Banco key-value embarcado mais moderno que `sled`. ACID, MVCC, zero unsafe code. Crescendo em popularidade.

### 10.7 Antes de `sqlx` e `diesel`

Queries SQL manuais com `postgres` crate (equivalente ao `lib/pq` do Go) ou `mysql`. Sem type safety — resultados eram `HashMap<String, Value>` ou structs com parsing manual. Muito verboso e propenso a erros em runtime.

---

## 11. Serialização e Formatos de Dados

### 11.1 `serde` (2014–presente)

O crate mais importante do ecossistema Rust, com bilhões de downloads. Framework de serialização/deserialização com suporte a qualquer formato via derive macros.

```rust
use serde::{Serialize, Deserialize};

#[derive(Debug, Serialize, Deserialize)]
struct Workspace {
    id: String,
    name: String,
    #[serde(skip_serializing_if = "Option::is_none")]
    description: Option<String>,
    #[serde(default)]
    active: bool,
    #[serde(rename = "created_at")]
    created: chrono::DateTime<chrono::Utc>,
}

// Serializar para JSON
let json = serde_json::to_string(&workspace)?;
// Deserializar de TOML
let config: Config = toml::from_str(&toml_str)?;
// Deserializar de MessagePack (binário)
let data: MyStruct = rmp_serde::from_slice(&bytes)?;
```

Serde é zero-cost — a serialização é monomorphizada em compile time para cada tipo, sem reflexão em runtime. Implementar `Serialize`/`Deserialize` manualmente ou via `#[derive(Serialize, Deserialize)]`.

**Formatos com suporte serde**: `serde_json`, `toml`, `serde_yaml`, `rmp-serde` (MessagePack), `bincode`, `ron` (Rusty Object Notation), `serde_cbor`, `postcard` (embedded), e dezenas de outros.

### 11.2 `serde_json` (2015–presente)

O mais usado dos backends serde. `Value` enum para JSON dinâmico, ou tipos estáticos via serde.

### 11.3 `toml` (2015–presente)

Parser TOML — o formato de configuração do `Cargo.toml`. Com serde, mapeia automaticamente para structs Rust.

### 11.4 `bincode` (2015–presente)

Serialização binária eficiente. Usado para serialização em disco, IPC de alta performance, caches.

---

## 12. Embedded e `no_std`

### O que é `no_std`

`#![no_std]` remove a stdlib do Rust — apenas `core` (sem alocação, sem OS, sem I/O) e opcionalmente `alloc` (heap allocation com alocador customizado). Permite rodar em microcontroladores sem SO.

```rust
#![no_std]
#![no_main]

use cortex_m_rt::entry;
use stm32f4xx_hal as hal;

#[entry]
fn main() -> ! {
    let peripherals = hal::pac::Peripherals::take().unwrap();
    let gpioa = peripherals.GPIOA.split();
    let mut led = gpioa.pa5.into_push_pull_output();

    loop {
        led.toggle();
        cortex_m::asm::delay(8_000_000); // delay ~1s em 8MHz
    }
}
```

### HAL (Hardware Abstraction Layer)

O ecossistema embedded Rust usa traits do `embedded-hal` como abstração sobre hardware:

- `SpiDevice` — interface SPI
- `I2c` — interface I2C
- `OutputPin`/`InputPin` — GPIO
- `DelayUs` — delays precisos

Drivers de sensores/displays implementam contra `embedded-hal` e funcionam em qualquer microcontrolador que implemente o HAL.

### Crates embedded importantes

**`cortex-m`** — suporte ao processador ARM Cortex-M (mais comum em MCUs).
**`embassy`** (2021–presente) — runtime async para embedded. Permite `async/await` em microcontroladores sem OS. Revolucionário: concorrência assíncrona sem threads, com footprint mínimo.
**`probe-rs`** — debug e flashing de MCUs via JTAG/SWD em Rust puro.
**`defmt`** — logging ultra-eficiente para embedded (format strings compiladas como índices numéricos, decodificadas no host).

---

## 13. WebAssembly

Rust é a linguagem de primeira classe para WASM de alto performance.

### `wasm-bindgen` (2018–presente)

Geração automática de bindings entre Rust (WASM) e JavaScript. Tipos JS nativos (String, Array, Promise) mapeados para tipos Rust.

```rust
use wasm_bindgen::prelude::*;

#[wasm_bindgen]
pub fn greet(name: &str) -> String {
    format!("Hello, {}!", name)
}

// JavaScript
// import init, { greet } from './pkg/mylib.js';
// await init();
// console.log(greet("World")); // "Hello, World!"
```

### `web-sys` e `js-sys` (2018–presente)

Bindings gerados automaticamente para todas as APIs do browser (DOM, fetch, WebGL, Canvas, WebAudio, etc.) e APIs JavaScript (Promise, Array, Object, etc.).

### `trunk` (2020–presente)

Build tool para WASM — compila, empacota e serve apps Rust/WASM com hot reloading. Equivalente ao Vite para projetos WASM.

### `wasm-pack` (2018–presente)

Empacota crates Rust como módulos npm. Permite publicar código Rust em npm e usar em projetos JavaScript/TypeScript.

---

## 14. Observabilidade e Testes

### 14.1 `tracing` (2019–presente)

Framework de structured logging, tracing e instrumentação. Substituto moderno do `log`. Spans com contexto, instrumentação de funções com `#[instrument]`.

```rust
use tracing::{info, warn, instrument, Span};

#[instrument(skip(password), fields(user_id))]
async fn login(username: &str, password: &str) -> Result<Token, AuthError> {
    info!("Login attempt");
    let user = db.find_user(username).await?;
    Span::current().record("user_id", user.id);
    // ...
    info!(duration_ms = %elapsed.as_millis(), "Login successful");
    Ok(token)
}
```

`tracing-subscriber` processa os eventos (para stdout, arquivo, OTLP). Integra com OpenTelemetry via `tracing-opentelemetry`.

### 14.2 `criterion` (2017–presente)

Framework de benchmarking estatístico. Mede performance com análise de regressão, histogramas, comparações entre versões.

```rust
use criterion::{criterion_group, criterion_main, Criterion};

fn benchmark_search(c: &mut Criterion) {
    c.bench_function("search 1000 docs", |b| {
        b.iter(|| db.search("godot signals"))
    });
}

criterion_group!(benches, benchmark_search);
criterion_main!(benches);
```

### 14.3 Testes em Rust

O sistema de testes é integrado no compilador:

```rust
// Testes unitários no mesmo arquivo
#[cfg(test)]
mod tests {
    use super::*;

    #[test]
    fn test_add() {
        assert_eq!(add(2, 3), 5);
    }

    #[test]
    #[should_panic(expected = "division by zero")]
    fn test_div_zero() {
        divide(1, 0);
    }

    #[tokio::test] // async test
    async fn test_fetch() {
        let result = fetch_data("http://example.com").await;
        assert!(result.is_ok());
    }
}
```

**`cargo test`** roda unit tests, integration tests (`tests/` dir) e doc tests (exemplos em doc comments).

### 14.4 `mockall` (2019–presente)

Geração de mocks automática via `#[automock]`:

```rust
#[cfg_attr(test, mockall::automock)]
trait Database {
    async fn get_user(&self, id: u64) -> Option<User>;
}

// No teste
let mut mock = MockDatabase::new();
mock.expect_get_user().with(eq(1)).returning(|_| Some(User { id: 1, name: "Test".into() }));
```

### 14.5 `proptest` e `quickcheck`

Property-based testing — gera inputs aleatórios automaticamente para encontrar edge cases:

```rust
use proptest::prelude::*;

proptest! {
    #[test]
    fn parse_serialize_roundtrip(s in "[a-z]{1,20}") {
        let parsed = parse(&s).unwrap();
        let serialized = serialize(&parsed);
        assert_eq!(s, serialized);
    }
}
```

---

## 15. Ferramentas de Desenvolvimento

### 15.1 `cargo` (2014–presente)

Build system, package manager, test runner, benchmark runner, documentation generator — tudo em um. Uma das maiores vantagens competitivas do Rust.

Comandos principais: `build`, `run`, `test`, `bench`, `check`, `fmt`, `clippy`, `doc`, `publish`, `add`, `remove`, `tree`, `audit`, `fix`, `vendor`.

### 15.2 `rustup` (2016–presente)

Toolchain manager. Instala e gerencia versões do compilador (stable, beta, nightly), targets para cross-compilation, e componentes (rustfmt, clippy, rust-analyzer, miri).

### 15.3 `rust-analyzer` (2020–presente)

Language Server para Rust. Alimenta IDE features em VSCode, Neovim, IntelliJ (via plugin), Emacs, e qualquer editor com suporte LSP. Substituto do `rls` (deprecated). Escrito em Rust.

### 15.4 `clippy` (2016–presente)

Linter. 700+ lints cobrindo performance, estilo, idiomas Rust, bugs comuns. Integrado ao cargo: `cargo clippy`. Lints por categoria: correctness, suspicious, style, complexity, perf, pedantic, nursery, restriction.

### 15.5 `rustfmt` (2015–presente)

Formatador automático de código. `cargo fmt`. Sem configuração de estilo — há um estilo oficial que `rustfmt` aplica. Fim de debates de formatação.

### 15.6 `miri` (2017–presente)

**Intérprete MIR** — executa Rust em um "ambiente virtual" que detecta undefined behavior, leituras de memória não-inicializada, violações de aliasing, race conditions. Essencial para validar código `unsafe`.

```bash
cargo +nightly miri test
# detecta: use-after-free, data races, UB em aritmética de ponteiros, etc.
```

### 15.7 `cargo-audit` (2018–presente)

Auditoria de segurança — verifica dependências contra o RustSec Advisory Database. `cargo audit` no CI.

### 15.8 `cargo-nextest` (2021–presente)

Test runner alternativo ao cargo test. Executa testes em paralelo por processo (mais isolamento), output muito mais limpo, retry automático de testes flaky, e muito mais rápido (2-3x).

### 15.9 `cargo-expand` (2018–presente)

Mostra o código gerado por macros após expansão. Essencial para debug de macros procedurais e derive.

### 15.10 `cross` (2017–presente)

Cross-compilation simples via Docker. `cross build --target aarch64-unknown-linux-gnu` funciona sem configurar toolchain manualmente.

---

## 16. Comparativo: Antes vs Hoje

| Área             | 2015 (Rust 1.0)        | 2018–2019                   | 2021–2022                      | 2024+                                 |
| ---------------- | ---------------------- | --------------------------- | ------------------------------ | ------------------------------------- |
| **Concorrência** | Threads OS + channels  | Futures 0.1 experimental    | async/await estável, Tokio 1.0 | async fn em traits, ecosystem maduro  |
| **Web Backend**  | iron (abandonado)      | actix-web, rocket (nightly) | axum 0.1, rocket stable        | axum, actix dominam; leptos SSR       |
| **Serialização** | manual + serde nightly | serde 1.0 stable            | universal                      | onipresente                           |
| **GUI**          | GTK bindings (frágil)  | electron sidecar            | tauri 1.0 RC, egui 0.1         | tauri 2.0 (mobile), ecosystem diverso |
| **Embedded**     | bare metal + unsafe    | embedded-hal                | embassy async embedded         | embassy maduro, probe-rs              |
| **WASM**         | Inexistente            | wasm-bindgen 0.1            | wasm-pack, trunk               | first-class target                    |
| **Banco**        | raw queries + rusqlite | diesel 1.0                  | sqlx async compile-time        | sqlx, diesel, sea-orm                 |
| **CLI**          | clap 2.x builder       | clap 3                      | ratatui, indicatif             | clap 4 derive, ratatui 0.26+          |
| **Tooling**      | clippy nightly         | clippy stable, rustfmt      | rust-analyzer                  | rust-analyzer maduro, miri            |
| **FFI**          | manual + bindgen       | cxx crate                   | cxx maduro, autocxx            | ecosystem de bindings rico            |
| **Erros**        | assustadores           | melhorando                  | muito bons                     | excepcionais                          |

### Adoção crescente

Rust entrou no **Linux Kernel** em 6.1 (dezembro 2022) — segunda linguagem suportada após C. **Microsoft** usa Rust em componentes do Windows Kernel, Azure, e serviços críticos. **Google** usa Rust no Android (bluetooth, keystore, virtualization), Chromium, e Fuchsia. **AWS** usa Rust em Firecracker (VMM do Lambda/Fargate) e Bottlerocket OS. **Meta** usa Rust no Sapling (SCM), diem (blockchain), e outras infras. **Cloudflare** usa Rust em produtos edge e Pingora (substituto do nginx).

Stack Overflow Developer Survey: Rust é a linguagem **mais amada/admirada** por 10 anos consecutivos (2016–2025), com 72% de admiration rate em 2025.

### A dicotomia fundamental que o Rust resolve

C/C++ oferecem controle total mas deixam o programador responsável pela segurança de memória e concorrência — erros são comuns, custosos, e às vezes fatais (buffer overflows em software crítico, data races em sistemas concorrentes). GC languages (Go, Java, Python) oferecem segurança automática mas pagam em latência de GC, overhead de runtime, e perda de controle. Rust oferece um terceiro caminho: o compilador verifica formalmente as invariantes de segurança via ownership e borrow checker — o programador tem controle total de memória e performance sem abrir mão de segurança.

O custo é a curva de aprendizado inicial: o borrow checker rejeita código que seria válido em outras linguagens, e entender por que requer internalizar o modelo de ownership. Mas esse custo é pago uma vez — depois, o compilador se torna um parceiro que previne bugs em vez de um obstáculo.
