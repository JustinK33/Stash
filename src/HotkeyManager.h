#pragma once

#include <QObject>
#include <QJsonObject>
#include <QString>
#include <functional>
#include <memory>

struct Hotkey {
  uint32_t keyCode = 49;
  bool fn = true;
  bool shift = false;
  bool ctrl = false;
  bool alt = false;
  bool cmd = false;

  static Hotkey defaultHotkey();
  QJsonObject toJson() const;
  static Hotkey fromJson(const QJsonObject &obj);
  QString toDisplayString() const;
  bool isLegacyFnZeroDefault() const;
};

class HotkeyManager : public QObject {
  Q_OBJECT
public:
  explicit HotkeyManager(QObject *parent = nullptr);
  ~HotkeyManager();

  struct Impl;

  bool start();
  void stop();

  bool setHotkey(const Hotkey &hotkey);
  Hotkey hotkey() const;

  bool captureNext(const std::function<void(const Hotkey &)> &callback);

signals:
  void activated();

private:
  std::unique_ptr<Impl> impl;
};
