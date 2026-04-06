**Русский** | [English](README_EN.md)

# billmind

Терминальное приложение для отслеживания подписок и регулярных платежей. Напомнит оплатить VPS, домен, Netflix и что угодно ещё — прямо в терминале.

## Возможности

- Управление подписками и разовыми платежами
- Три стадии оповещений: мягкое → срочное → критическое
- Системные нотификации (macOS, Linux, Windows) с кнопкой открытия TUI
- Фоновый демон — проверяет платежи каждый час
- Тихие часы (22:00–08:00) — не беспокоит ночью
- Vim-навигация (j/k, gg/G, /, dd)
- Поиск и фильтрация по тегам
- Undo для любого действия
- Автоматические бэкапы (последние 10)
- Поддержка русского и английского языков

## Требования

- Go 1.24+
- `terminal-notifier` (macOS, рекомендуется для persistent-нотификаций)

## Установка

### Homebrew (macOS / Linux)

```bash
brew tap curkan/public
brew install billmind
```

### Сборка из исходников

```bash
git clone https://github.com/curkan/billmind.git
cd billmind
go build -o billmind ./cmd/billmind
```

### Скачать бинарник

Готовые бинарники для всех платформ — на странице [Releases](https://github.com/curkan/billmind/releases).

### macOS — установка terminal-notifier (рекомендуется)

```bash
brew install terminal-notifier
```

Без `terminal-notifier` нотификации будут работать через `osascript`, но исчезают через 5 секунд и не имеют кнопки действия. С `terminal-notifier` — persistent-нотификации со звуком и кнопкой "Показать".

## Запуск

```bash
# Открыть TUI
./billmind

# Запустить демон вручную (проверить и отправить оповещения)
./billmind daemon

# Установить демон в планировщик ОС (запуск каждый час)
./billmind install

# Удалить демон из планировщика
./billmind uninstall
```

## Управление

### Навигация

| Клавиша | Действие |
|---------|----------|
| `j` / `↓` | Вниз |
| `k` / `↑` | Вверх |
| `gg` | В начало списка |
| `G` | В конец списка |

### Действия

| Клавиша | Действие |
|---------|----------|
| `a` | Добавить напоминание (wizard) |
| `e` | Редактировать выбранное |
| `dd` | Удалить (двойное нажатие + подтверждение) |
| `Space` | Отметить как оплаченное |
| `o` | Открыть URL в браузере |
| `u` | Отменить последнее действие |
| `/` | Поиск по названию, тегам, URL |
| `f` | Фильтр по тегам |
| `?` | Справка |
| `q` | Выход |

### В формах (wizard / редактирование)

| Клавиша | Действие |
|---------|----------|
| `Tab` / `Ctrl+J` | Следующее поле |
| `Shift+Tab` / `Ctrl+K` | Предыдущее поле |
| `Enter` | Сохранить |
| `Esc` | Отмена |
| `Space` | Переключить чекбокс |
| `h` / `l` | Выбор интервала (← / →) |

## Wizard — добавление напоминания

Создание нового напоминания проходит в 4 шага:

1. **Информация** — название, URL (необязательно), теги через запятую
2. **Расписание** — интервал (еженедельно / ежемесячно / ежегодно / своё / разовый), дата оплаты, за сколько дней напоминать
3. **Оповещения** — системные нотификации (macOS/Linux/Windows), ntfy.sh (push на телефон)
4. **Подтверждение** — сводка, сохранение

## Система оповещений

### Три стадии (триада)

Каждое напоминание проходит через три стадии нотификаций за один цикл оплаты:

| Стадия | Когда | Пример |
|--------|-------|--------|
| **Мягкое** (soft) | За N дней до оплаты | "Hetzner VPS — payment in 3 day(s) (May 01)" |
| **Срочное** (urgent) | В день оплаты | "Hetzner VPS — payment today!" |
| **Критическое** (critical) | После просрочки | "Hetzner VPS — overdue by 2 day(s)" |

- Каждая стадия отправляется **ровно один раз** — никакого спама
- Если несколько платежей на одну дату — одно пакетное уведомление
- После оплаты стадии сбрасываются для следующего цикла

### Тихие часы

По умолчанию 22:00–08:00. Демон не отправляет нотификации в это время. При пробуждении — догоняет пропущенное.

Настраивается в `~/.config/billmind/data.json`:

```json
{
  "settings": {
    "quiet_hours_start": 22,
    "quiet_hours_end": 8
  }
}
```

### Catch-up после сна

Если ноутбук был в спящем режиме и пропустил несколько стадий — демон отправит только **самую актуальную**. Не будет слать "оплати через 3 дня", если уже просрочено.

### ntfy.sh — push-уведомления на телефон

[ntfy.sh](https://ntfy.sh/) позволяет получать push-уведомления на телефон без регистрации и API-ключей.

**Настройка (одноразовая):**

1. Установите приложение ntfy на телефон ([iOS](https://apps.apple.com/app/ntfy/id1625396347) / [Android](https://play.google.com/store/apps/details?id=io.heckel.ntfy))
2. Откройте приложение и подпишитесь на топик — придумайте уникальное имя, например `billmind-ivan-2026` (это как личный канал, знает только тот, кто подписан)
3. В billmind нажмите `s` (настройки) → `Tab` → введите тот же топик → `Esc`
4. При добавлении или редактировании напоминания включите ntfy-переключатель (шаг 3 wizard или `e` → ntfy toggle)

**Готово.** Теперь при каждой стадии оповещения придёт push на телефон со звуком.

**Проверка:**

```bash
# Отправить тестовое уведомление вручную
curl -d "Test notification from billmind" ntfy.sh/ВАШ_ТОПИК
```

Если push пришёл на телефон — всё настроено правильно.

### Платформы

| Платформа | Системные нотификации | Push (ntfy.sh) | Планировщик |
|-----------|----------------------|----------------|-------------|
| **macOS** | `terminal-notifier` (persistent + звук + кнопка) / `osascript` (fallback) | HTTP POST | launchd (`~/Library/LaunchAgents/`) |
| **Linux** | D-Bus notify-send | HTTP POST | systemd user timer (`~/.config/systemd/user/`) |
| **Windows** | Windows Toast | HTTP POST | schtasks (Task Scheduler) |

## Хранение данных

Все данные хранятся в `~/.config/billmind/`:

```
~/.config/billmind/
├── data.json           # Напоминания и настройки
├── daemon.log          # Логи демона
└── backups/            # Автоматические бэкапы (до 10 штук)
    ├── data_2026-04-06_120000.json
    └── ...
```

### Структура data.json

```json
{
  "reminders": [
    {
      "id": "550e8400-e29b-41d4-a716-446655440000",
      "name": "Hetzner VPS",
      "url": "https://console.hetzner.cloud/billing",
      "tags": ["work", "vps"],
      "interval": "monthly",
      "next_due": "2026-05-01T00:00:00Z",
      "remind_days_before": 3,
      "notifications": {
        "macos": true,
        "ntfy": true
      },
      "notify_stage": 0,
      "paid_at": null
    }
  ],
  "settings": {
    "language": "ru",
    "ntfy_topic": "billmind-myname123",
    "quiet_hours_start": 22,
    "quiet_hours_end": 8
  }
}
```

### Интервалы оплаты

| Значение | Описание |
|----------|----------|
| `weekly` | Еженедельно |
| `monthly` | Ежемесячно |
| `yearly` | Ежегодно |
| `once` | Разовый платёж (удаляется после оплаты) |
| `custom` | Свой интервал в днях (`custom_days`) |

### Стадии нотификаций (notify_stage)

| Значение | Стадия | Описание |
|----------|--------|----------|
| `0` | none | Ещё не оповещали |
| `1` | soft | Мягкое отправлено |
| `2` | urgent | Срочное отправлено |
| `3` | critical | Критическое отправлено |

## Разработка

```bash
# Запуск в режиме разработки
go run ./cmd/billmind

# Тесты с race detector
go test ./... -race

# Тесты конкретного пакета
go test ./internal/daemon/... -race -v

# Сборка
go build -o billmind ./cmd/billmind
```

### Архитектура

Проект построен на MVU (Model-View-Update) архитектуре Elm:

```
cmd/billmind/main.go          # Entry point + CLI субкоманды
internal/
  ui/                          # TUI (Bubbletea v2)
    model.go                   # Состояние приложения
    update.go                  # Обработка сообщений
    view.go                    # Рендеринг
    handlers_*.go              # Обработчики по экранам
    wizard.go                  # Wizard добавления
  daemon/                      # Фоновый демон
    daemon.go                  # Оркестратор Run()
    notify.go                  # Группировка и отправка
    quiethours.go              # Тихие часы
  domain/                      # Доменные модели
    models.go                  # Reminder, Interval, NotifyStage
  storage/                     # Персистентность (JSON)
  platform/                    # Платформенные абстракции
    darwin.go                  # macOS
    linux.go                   # Linux
    windows.go                 # Windows
    fallback.go                # Fallback
  i18n/                        # Интернационализация (ru/en)
```

## Решение проблем

### Нотификации исчезают мгновенно (macOS)

Установите `terminal-notifier`:

```bash
brew install terminal-notifier
```

Или измените тип нотификаций в System Settings → Notifications → Script Editor → Alerts.

### Демон не запускается

Проверьте логи:

```bash
cat ~/.config/billmind/daemon.log
```

Проверьте статус в планировщике:

```bash
# macOS
launchctl list | grep billmind

# Linux
systemctl --user status com.billmind.daemon.timer
```

### Нотификации не приходят

1. Проверьте, что `notifications.macos: true` в напоминании
2. Проверьте, что сейчас не тихие часы (22:00–08:00)
3. Проверьте `notify_stage` — если уже `3`, стадии исчерпаны до следующего цикла
4. Запустите демон вручную: `./billmind daemon`

### Данные пропали

Проверьте бэкапы:

```bash
ls ~/.config/billmind/backups/
```

Скопируйте последний бэкап:

```bash
cp ~/.config/billmind/backups/data_ПОСЛЕДНИЙ.json ~/.config/billmind/data.json
```

## Релизы

Инструкция по публикации новых версий — [RELEASING.md](RELEASING.md).

## Лицензия

MIT License

## Поддержка

Если вы столкнулись с проблемами, создайте issue:
https://github.com/curkan/billmind/issues
