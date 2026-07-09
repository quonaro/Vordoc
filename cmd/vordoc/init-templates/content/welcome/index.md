---
title: Полноценный пример Vordoc
description: Демонстрация нескольких документаций, кастомных компонентов, ссылок, якорей и авторизации
order: 0
---

# Полноценный пример Vordoc

Этот проект создан командой `vordoc init` и показывает, как организовать несколько независимых документаций в одном сайте.

## Документации

- **[Компоненты](components/index.md)** — встроенные виджеты: callouts, галереи, изображения, диаграммы, блоки кода.
- **[Ссылки и якори](links/index.md)** — внутренние переходы между документациями, навигация по заголовкам, внешние ссылки.
- **[Закрытый раздел](members/index.md)** — документация `members`, защищённая паролем `member`.
- **[Админ-раздел](admin/index.md)** — документация `admin`, защищённая паролем `admin`, с публичной подстраницей внутри.

## Структура контента

```text
content/
├── config.yaml
├── text.json
├── welcome/
│   ├── config.yaml
│   └── index.md
├── components/
│   ├── config.yaml
│   ├── index.md
│   └── public/
│       ├── sample.svg
│       └── notes.txt
├── links/
│   ├── config.yaml
│   └── index.md
├── members/        # пароль: member
│   ├── config.yaml
│   ├── index.md
│   └── secret.md
└── admin/          # пароль: admin
    ├── config.yaml
    ├── index.md
    └── public/     # публичный override
        ├── config.yaml
        └── info.md
```

## Навигация

- [Компоненты](components/index.md)
- [Ссылки и якори](links/index.md)
- [Закрытый раздел](members/index.md)
- [Админ-раздел](admin/index.md)
- [Публичная страница внутри админ-раздела](admin/public/info.md)
