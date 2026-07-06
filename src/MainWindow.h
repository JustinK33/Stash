#pragma once

#include <QMainWindow>
#include <QPointer>
#include <QVector>

#include "DataStore.h"
#include "HotkeyManager.h"

class QTextEdit;
class QLabel;
class QVBoxLayout;
class QClipboard;
class KeybindPill;
class QLineEdit;
class QComboBox;
class QSystemTrayIcon;
class QMenu;
class QPropertyAnimation;
class QFocusEvent;

class MainWindow : public QMainWindow {
  Q_OBJECT
public:
  explicit MainWindow(QWidget *parent = nullptr);
  ~MainWindow() override = default;

protected:
  bool eventFilter(QObject *obj, QEvent *event) override;
  void focusOutEvent(QFocusEvent *event) override;

private slots:
  void saveSnippetFromInput();
  void toggleVisibility();
  void beginHotkeyCapture();
  void saveData();

private:
  void buildUi();
  void applyTheme();
  void refreshSnippetsUi();
  void loadData();
  void createTray();
  void animateShow();
  void animateHide();
  void animateFadeIn(QWidget *widget, int durationMs);
  void setStatus(const QString &text);

  DataStore store;
  QVector<Note> notes;
  QVector<ClipItem> clips;
  Hotkey hotkey;
  Settings settings;

  QTextEdit *noteInput = nullptr;
  QLabel *clipsCountLabel = nullptr;
  QLabel *statusLabel = nullptr;
  QWidget *topBar = nullptr;
  KeybindPill *keybindPill = nullptr;
  QVBoxLayout *clipsListLayout = nullptr;
  QString lastAddedClipText;
  QClipboard *clipboard = nullptr;
  HotkeyManager *hotkeyManager = nullptr;
  QSystemTrayIcon *trayIcon = nullptr;
  QMenu *trayMenu = nullptr;
  QPointer<QPropertyAnimation> opacityAnimation;
  bool dragging = false;
  QPoint dragStart;
};
