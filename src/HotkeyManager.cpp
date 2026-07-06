#include "HotkeyManager.h"

#include <QMetaObject>
#include <QStringList>

#ifdef Q_OS_MAC
#include <ApplicationServices/ApplicationServices.h>
#include <Carbon/Carbon.h>
#endif

namespace {
constexpr uint32_t hotkeySignature = 'QKNT';
constexpr uint32_t hotkeyIdentifier = 1;

QString keyCodeToString(uint32_t keyCode) {
#ifdef Q_OS_MAC
  struct Entry {
    uint32_t code;
    const char *name;
  };

  static const Entry entries[] = {
      {kVK_Space, "Space"},
      {kVK_Return, "Enter"},
      {kVK_Tab, "Tab"},
      {kVK_Escape, "Esc"},
      {kVK_Delete, "Backspace"},
      {kVK_ForwardDelete, "Delete"},
      {kVK_LeftArrow, "Left"},
      {kVK_RightArrow, "Right"},
      {kVK_UpArrow, "Up"},
      {kVK_DownArrow, "Down"},
      {kVK_ANSI_A, "A"},
      {kVK_ANSI_B, "B"},
      {kVK_ANSI_C, "C"},
      {kVK_ANSI_D, "D"},
      {kVK_ANSI_E, "E"},
      {kVK_ANSI_F, "F"},
      {kVK_ANSI_G, "G"},
      {kVK_ANSI_H, "H"},
      {kVK_ANSI_I, "I"},
      {kVK_ANSI_J, "J"},
      {kVK_ANSI_K, "K"},
      {kVK_ANSI_L, "L"},
      {kVK_ANSI_M, "M"},
      {kVK_ANSI_N, "N"},
      {kVK_ANSI_O, "O"},
      {kVK_ANSI_P, "P"},
      {kVK_ANSI_Q, "Q"},
      {kVK_ANSI_R, "R"},
      {kVK_ANSI_S, "S"},
      {kVK_ANSI_T, "T"},
      {kVK_ANSI_U, "U"},
      {kVK_ANSI_V, "V"},
      {kVK_ANSI_W, "W"},
      {kVK_ANSI_X, "X"},
      {kVK_ANSI_Y, "Y"},
      {kVK_ANSI_Z, "Z"},
      {kVK_ANSI_0, "0"},
      {kVK_ANSI_1, "1"},
      {kVK_ANSI_2, "2"},
      {kVK_ANSI_3, "3"},
      {kVK_ANSI_4, "4"},
      {kVK_ANSI_5, "5"},
      {kVK_ANSI_6, "6"},
      {kVK_ANSI_7, "7"},
      {kVK_ANSI_8, "8"},
      {kVK_ANSI_9, "9"},
  };

  for (const auto &entry : entries) {
    if (entry.code == keyCode) {
      return QString::fromLatin1(entry.name);
    }
  }
#endif
  return QString("Key %1").arg(keyCode);
}
} // namespace

Hotkey Hotkey::defaultHotkey() {
  Hotkey hotkey;
#ifdef Q_OS_MAC
  hotkey.keyCode = kVK_ANSI_0;
#else
  hotkey.keyCode = 0;
#endif
  hotkey.fn = false;
  hotkey.shift = false;
  hotkey.ctrl = true;
  hotkey.alt = true;
  hotkey.cmd = false;
  return hotkey;
}

QJsonObject Hotkey::toJson() const {
  QJsonObject obj;
  obj["keyCode"] = static_cast<int>(keyCode);
  obj["fn"] = fn;
  obj["shift"] = shift;
  obj["ctrl"] = ctrl;
  obj["alt"] = alt;
  obj["cmd"] = cmd;
  obj["display"] = toDisplayString();
  return obj;
}

Hotkey Hotkey::fromJson(const QJsonObject &obj) {
  Hotkey hotkey = Hotkey::defaultHotkey();
  if (obj.isEmpty()) {
    return hotkey;
  }

  hotkey.keyCode = static_cast<uint32_t>(
      obj.value("keyCode").toInt(static_cast<int>(hotkey.keyCode)));
  hotkey.fn = obj.value("fn").toBool(hotkey.fn);
  hotkey.shift = obj.value("shift").toBool(hotkey.shift);
  hotkey.ctrl = obj.value("ctrl").toBool(hotkey.ctrl);
  hotkey.alt = obj.value("alt").toBool(hotkey.alt);
  hotkey.cmd = obj.value("cmd").toBool(hotkey.cmd);
  return hotkey;
}

QString Hotkey::toDisplayString() const {
  QStringList parts;
  if (fn) {
    parts << "fn";
  }
  if (cmd) {
    parts << "Cmd";
  }
  if (alt) {
    parts << "Option";
  }
  if (ctrl) {
    parts << "Control";
  }
  if (shift) {
    parts << "Shift";
  }
  parts << keyCodeToString(keyCode);
  return parts.join(" + ");
}

bool Hotkey::isLegacyFnZeroDefault() const {
#ifdef Q_OS_MAC
  return keyCode == kVK_ANSI_0 && fn && !shift && !ctrl && !alt && !cmd;
#else
  return false;
#endif
}

struct HotkeyManager::Impl {
  explicit Impl(HotkeyManager *owner) : owner(owner) {}

  HotkeyManager *owner = nullptr;
  Hotkey current = Hotkey::defaultHotkey();
  std::function<void(const Hotkey &)> captureCallback;

#ifdef Q_OS_MAC
  CFMachPortRef eventTap = nullptr;
  CFRunLoopSourceRef runLoopSource = nullptr;
  EventHotKeyRef carbonHotkey = nullptr;
  EventHandlerRef carbonHandler = nullptr;
  bool started = false;

  bool matches(CGEventFlags flags, uint32_t keyCode) const {
    if (keyCode != current.keyCode) {
      return false;
    }

    CGEventFlags required = 0;
    if (current.shift) {
      required |= kCGEventFlagMaskShift;
    }
    if (current.ctrl) {
      required |= kCGEventFlagMaskControl;
    }
    if (current.alt) {
      required |= kCGEventFlagMaskAlternate;
    }
    if (current.cmd) {
      required |= kCGEventFlagMaskCommand;
    }
    if (current.fn) {
      required |= kCGEventFlagMaskSecondaryFn;
    }

    CGEventFlags relevant = kCGEventFlagMaskShift | kCGEventFlagMaskControl |
                            kCGEventFlagMaskAlternate | kCGEventFlagMaskCommand |
                            kCGEventFlagMaskSecondaryFn;

    return (flags & relevant) == required;
  }

  Hotkey fromFlags(CGEventFlags flags, uint32_t keyCode) const {
    Hotkey hotkey = current;
    hotkey.keyCode = keyCode;
    hotkey.shift = flags & kCGEventFlagMaskShift;
    hotkey.ctrl = flags & kCGEventFlagMaskControl;
    hotkey.alt = flags & kCGEventFlagMaskAlternate;
    hotkey.cmd = flags & kCGEventFlagMaskCommand;
    hotkey.fn = flags & kCGEventFlagMaskSecondaryFn;
    return hotkey;
  }

  uint32_t carbonModifiers() const {
    uint32_t modifiers = 0;
    if (current.shift) {
      modifiers |= shiftKey;
    }
    if (current.ctrl) {
      modifiers |= controlKey;
    }
    if (current.alt) {
      modifiers |= optionKey;
    }
    if (current.cmd) {
      modifiers |= cmdKey;
    }
    return modifiers;
  }
#endif
};

#ifdef Q_OS_MAC
static OSStatus carbonHotkeyCallback(EventHandlerCallRef, EventRef event,
                                     void *refcon) {
  auto *impl = static_cast<HotkeyManager::Impl *>(refcon);
  if (!impl || !impl->owner ||
      GetEventClass(event) != kEventClassKeyboard ||
      GetEventKind(event) != kEventHotKeyPressed) {
    return noErr;
  }

  QMetaObject::invokeMethod(
      impl->owner, [impl]() { emit impl->owner->activated(); },
      Qt::QueuedConnection);
  return noErr;
}

static CGEventRef hotkeyCallback(CGEventTapProxy, CGEventType type,
                                 CGEventRef event, void *refcon) {
  auto *impl = static_cast<HotkeyManager::Impl *>(refcon);
  if (!impl || !impl->owner) {
    return event;
  }

  if (type == kCGEventTapDisabledByTimeout ||
      type == kCGEventTapDisabledByUserInput) {
    if (impl->eventTap) {
      CGEventTapEnable(impl->eventTap, true);
    }
    return event;
  }

  if (type != kCGEventKeyDown) {
    return event;
  }

  uint32_t keyCode =
      static_cast<uint32_t>(CGEventGetIntegerValueField(
          event, kCGKeyboardEventKeycode));
  CGEventFlags flags = CGEventGetFlags(event);

  if (impl->captureCallback) {
    Hotkey captured = impl->fromFlags(flags, keyCode);
    auto callback = impl->captureCallback;
    impl->captureCallback = nullptr;
    QMetaObject::invokeMethod(
        impl->owner, [callback, captured]() { callback(captured); },
        Qt::QueuedConnection);
    return event;
  }

  if (impl->matches(flags, keyCode)) {
    QMetaObject::invokeMethod(
        impl->owner, [impl]() { emit impl->owner->activated(); },
        Qt::QueuedConnection);
  }

  return event;
}
#endif

HotkeyManager::HotkeyManager(QObject *parent)
    : QObject(parent), impl(std::make_unique<Impl>(this)) {}

HotkeyManager::~HotkeyManager() {
  stop();
}

bool HotkeyManager::start() {
#ifdef Q_OS_MAC
  if (impl->started) {
    return true;
  }

  if (!impl->current.fn) {
    EventTypeSpec eventType;
    eventType.eventClass = kEventClassKeyboard;
    eventType.eventKind = kEventHotKeyPressed;

    if (InstallEventHandler(GetApplicationEventTarget(), carbonHotkeyCallback,
                            1, &eventType, impl.get(),
                            &impl->carbonHandler) != noErr) {
      return false;
    }

    EventHotKeyID hotkeyId;
    hotkeyId.signature = hotkeySignature;
    hotkeyId.id = hotkeyIdentifier;
    if (RegisterEventHotKey(impl->current.keyCode, impl->carbonModifiers(),
                            hotkeyId, GetApplicationEventTarget(), 0,
                            &impl->carbonHotkey) != noErr) {
      if (impl->carbonHandler) {
        RemoveEventHandler(impl->carbonHandler);
        impl->carbonHandler = nullptr;
      }
      return false;
    }

    impl->started = true;
    return true;
  }

  CGEventMask mask = CGEventMaskBit(kCGEventKeyDown);
  impl->eventTap = CGEventTapCreate(kCGSessionEventTap, kCGHeadInsertEventTap,
                                    kCGEventTapOptionListenOnly, mask,
                                    hotkeyCallback, impl.get());

  if (!impl->eventTap) {
    return false;
  }

  impl->runLoopSource =
      CFMachPortCreateRunLoopSource(kCFAllocatorDefault, impl->eventTap, 0);
  CFRunLoopAddSource(CFRunLoopGetCurrent(), impl->runLoopSource,
                     kCFRunLoopCommonModes);
  CGEventTapEnable(impl->eventTap, true);
  impl->started = true;
  return true;
#else
  return false;
#endif
}

void HotkeyManager::stop() {
#ifdef Q_OS_MAC
  if (!impl->started) {
    return;
  }

  if (impl->carbonHotkey) {
    UnregisterEventHotKey(impl->carbonHotkey);
    impl->carbonHotkey = nullptr;
  }

  if (impl->carbonHandler) {
    RemoveEventHandler(impl->carbonHandler);
    impl->carbonHandler = nullptr;
  }

  if (impl->runLoopSource) {
    CFRunLoopRemoveSource(CFRunLoopGetCurrent(), impl->runLoopSource,
                          kCFRunLoopCommonModes);
    CFRelease(impl->runLoopSource);
    impl->runLoopSource = nullptr;
  }

  if (impl->eventTap) {
    CFMachPortInvalidate(impl->eventTap);
    CFRelease(impl->eventTap);
    impl->eventTap = nullptr;
  }

  impl->started = false;
#endif
}

bool HotkeyManager::setHotkey(const Hotkey &hotkey) {
  const bool wasStarted = impl->started;
  if (wasStarted) {
    stop();
  }
  impl->current = hotkey;
  if (wasStarted) {
    return start();
  }
  return true;
}

Hotkey HotkeyManager::hotkey() const {
  return impl->current;
}

bool HotkeyManager::captureNext(
    const std::function<void(const Hotkey &)> &callback) {
  impl->captureCallback = callback;
#ifdef Q_OS_MAC
  if (!callback) {
    if (!impl->current.fn) {
      if (impl->runLoopSource) {
        CFRunLoopRemoveSource(CFRunLoopGetCurrent(), impl->runLoopSource,
                              kCFRunLoopCommonModes);
        CFRelease(impl->runLoopSource);
        impl->runLoopSource = nullptr;
      }
      if (impl->eventTap) {
        CFMachPortInvalidate(impl->eventTap);
        CFRelease(impl->eventTap);
        impl->eventTap = nullptr;
      }
    }
    return true;
  }
  if (impl->eventTap) {
    return true;
  }

  CGEventMask mask = CGEventMaskBit(kCGEventKeyDown);
  impl->eventTap = CGEventTapCreate(kCGSessionEventTap, kCGHeadInsertEventTap,
                                    kCGEventTapOptionListenOnly, mask,
                                    hotkeyCallback, impl.get());
  if (!impl->eventTap) {
    impl->captureCallback = nullptr;
    return false;
  }

  impl->runLoopSource =
      CFMachPortCreateRunLoopSource(kCFAllocatorDefault, impl->eventTap, 0);
  CFRunLoopAddSource(CFRunLoopGetCurrent(), impl->runLoopSource,
                     kCFRunLoopCommonModes);
  CGEventTapEnable(impl->eventTap, true);
  return true;
#else
  return false;
#endif
}
