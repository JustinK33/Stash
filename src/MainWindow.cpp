#include "MainWindow.h"

#include "widgets/ClipItemWidget.h"
#include "widgets/KeybindPill.h"

#include <QAbstractAnimation>
#include <QApplication>
#include <QClipboard>
#include <QDateTime>
#include <QDialog>
#include <QFrame>
#include <QGraphicsOpacityEffect>
#include <QGuiApplication>
#include <QHBoxLayout>
#include <QIcon>
#include <QLabel>
#include <QMenu>
#include <QMouseEvent>
#include <QPainter>
#include <QPixmap>
#include <QPropertyAnimation>
#include <QPushButton>
#include <QScrollArea>
#include <QShortcut>
#include <QSystemTrayIcon>
#include <QTextEdit>
#include <QToolButton>
#include <QVBoxLayout>

#include <algorithm>

namespace {
void clearLayout(QLayout *layout) {
  while (auto *item = layout->takeAt(0)) {
    if (item->widget()) {
      item->widget()->deleteLater();
    }
    delete item;
  }
}

ClipItem createClip(const QString &text) {
  ClipItem clip;
  clip.text = text;
  clip.captured = QDateTime::currentDateTime();
  return clip;
}

QIcon createTrayIcon() {
  QPixmap pixmap(32, 32);
  pixmap.fill(Qt::transparent);

  QPainter painter(&pixmap);
  painter.setRenderHint(QPainter::Antialiasing);
  painter.setBrush(QColor("#2f6fed"));
  painter.setPen(Qt::NoPen);
  painter.drawRoundedRect(QRectF(4, 4, 24, 24), 6, 6);
  painter.setPen(QPen(Qt::white, 3, Qt::SolidLine, Qt::RoundCap));
  painter.drawLine(QPointF(11, 12), QPointF(21, 12));
  painter.drawLine(QPointF(11, 18), QPointF(18, 18));
  return QIcon(pixmap);
}
} // namespace

MainWindow::MainWindow(QWidget *parent) : QMainWindow(parent) {
  setWindowTitle("QuickNote");
  resize(560, 620);
  setMinimumSize(420, 520);
  setWindowFlags(Qt::FramelessWindowHint | Qt::WindowStaysOnTopHint | Qt::Tool);
  setAttribute(Qt::WA_TranslucentBackground);

  loadData();
  buildUi();
  applyTheme();
  createTray();

  keybindPill->setText(hotkey.toDisplayString());

  clipboard = QGuiApplication::clipboard();
  hotkeyManager = new HotkeyManager(this);
  hotkeyManager->setHotkey(hotkey);
  connect(hotkeyManager, &HotkeyManager::activated, this,
          &MainWindow::toggleVisibility);
  hotkeyManager->start();

  refreshSnippetsUi();
  setStatus("Ready");
}

void MainWindow::buildUi() {
  auto *root = new QWidget(this);
  root->setObjectName("Window");
  setCentralWidget(root);

  auto *rootLayout = new QVBoxLayout(root);
  rootLayout->setContentsMargins(10, 10, 10, 10);
  rootLayout->setSpacing(0);

  auto *frame = new QFrame(root);
  frame->setObjectName("Frame");
  rootLayout->addWidget(frame);

  auto *frameLayout = new QVBoxLayout(frame);
  frameLayout->setContentsMargins(0, 0, 0, 0);
  frameLayout->setSpacing(0);

  topBar = new QWidget(frame);
  topBar->setObjectName("TopBar");
  topBar->setFixedHeight(54);
  topBar->installEventFilter(this);
  frameLayout->addWidget(topBar);

  auto *topLayout = new QHBoxLayout(topBar);
  topLayout->setContentsMargins(18, 0, 18, 0);
  topLayout->setSpacing(10);

  auto *title = new QLabel("QuickNote", topBar);
  title->setObjectName("Title");

  keybindPill = new KeybindPill(topBar);
  keybindPill->setFixedHeight(28);
  connect(keybindPill, &KeybindPill::editClicked, this,
          &MainWindow::beginHotkeyCapture);

  topLayout->addWidget(title, 0, Qt::AlignLeft);
  topLayout->addStretch(1);
  topLayout->addWidget(keybindPill, 0, Qt::AlignRight);

  auto *content = new QWidget(frame);
  content->setObjectName("Content");
  frameLayout->addWidget(content, 1);

  auto *contentLayout = new QVBoxLayout(content);
  contentLayout->setContentsMargins(18, 18, 18, 14);
  contentLayout->setSpacing(14);

  noteInput = new QTextEdit(content);
  noteInput->setObjectName("SnippetInput");
  noteInput->setPlaceholderText("Paste text here to save it");
  noteInput->setMinimumHeight(130);

  auto *actionsRow = new QHBoxLayout();
  actionsRow->setSpacing(10);

  auto *saveButton = new QPushButton("Save", content);
  saveButton->setObjectName("PrimaryButton");
  saveButton->setDefault(true);

  auto *clearInputButton = new QPushButton("Clear", content);
  clearInputButton->setObjectName("SecondaryButton");

  auto *clearAllButton = new QPushButton("Clear All", content);
  clearAllButton->setObjectName("SecondaryButton");

  actionsRow->addWidget(saveButton);
  actionsRow->addWidget(clearInputButton);
  actionsRow->addStretch(1);
  actionsRow->addWidget(clearAllButton);

  auto *listHeader = new QHBoxLayout();
  listHeader->setSpacing(8);

  auto *savedLabel = new QLabel("Saved snippets", content);
  savedLabel->setObjectName("SectionLabel");
  clipsCountLabel = new QLabel(content);
  clipsCountLabel->setObjectName("SectionValue");

  listHeader->addWidget(savedLabel);
  listHeader->addStretch(1);
  listHeader->addWidget(clipsCountLabel);

  auto *clipsListContainer = new QWidget(content);
  clipsListContainer->setObjectName("SnippetsList");
  clipsListLayout = new QVBoxLayout(clipsListContainer);
  clipsListLayout->setContentsMargins(0, 0, 0, 0);
  clipsListLayout->setSpacing(8);

  auto *clipsScroll = new QScrollArea(content);
  clipsScroll->setObjectName("SnippetsScroll");
  clipsScroll->setWidgetResizable(true);
  clipsScroll->setWidget(clipsListContainer);

  contentLayout->addWidget(noteInput);
  contentLayout->addLayout(actionsRow);
  contentLayout->addLayout(listHeader);
  contentLayout->addWidget(clipsScroll, 1);

  auto *footer = new QWidget(frame);
  footer->setObjectName("Footer");
  footer->setFixedHeight(42);
  frameLayout->addWidget(footer);

  auto *footerLayout = new QHBoxLayout(footer);
  footerLayout->setContentsMargins(18, 0, 18, 0);

  statusLabel = new QLabel("Ready", footer);
  statusLabel->setObjectName("StatusLabel");
  footerLayout->addWidget(statusLabel);
  footerLayout->addStretch(1);

  connect(saveButton, &QPushButton::clicked, this,
          &MainWindow::saveSnippetFromInput);
  connect(clearInputButton, &QPushButton::clicked, noteInput,
          &QTextEdit::clear);
  connect(clearAllButton, &QPushButton::clicked, this, [this]() {
    clips.clear();
    refreshSnippetsUi();
    saveData();
  });

  auto *saveShortcut =
      new QShortcut(QKeySequence(Qt::CTRL | Qt::Key_Return), this);
  connect(saveShortcut, &QShortcut::activated, this,
          &MainWindow::saveSnippetFromInput);

  auto *saveShortcutMeta =
      new QShortcut(QKeySequence(Qt::META | Qt::Key_Return), this);
  connect(saveShortcutMeta, &QShortcut::activated, this,
          &MainWindow::saveSnippetFromInput);
}

void MainWindow::applyTheme() {
  QFont baseFont("SF Pro Text", 12);
  QApplication::setFont(baseFont);

  const QString style =
      "QWidget { color: #20242a; }"
      "QFrame#Frame {"
      "  background: #f8f9fb;"
      "  border: 1px solid #d9dde5;"
      "  border-radius: 8px;"
      "}"
      "QWidget#TopBar, QWidget#Footer { background: #ffffff; }"
      "QWidget#TopBar { border-bottom: 1px solid #e3e6eb; }"
      "QWidget#Footer { border-top: 1px solid #e3e6eb; }"
      "QTextEdit#SnippetInput {"
      "  background: #ffffff;"
      "  border: 1px solid #cfd5df;"
      "  border-radius: 8px;"
      "  padding: 10px;"
      "  color: #20242a;"
      "  selection-background-color: #2f6fed;"
      "}"
      "QWidget#ClipItem {"
      "  background: #ffffff;"
      "  border: 1px solid #dfe3ea;"
      "  border-radius: 8px;"
      "}"
      "QPushButton#PrimaryButton {"
      "  background: #2f6fed;"
      "  border: 1px solid #245fd0;"
      "  border-radius: 7px;"
      "  color: #ffffff;"
      "  padding: 7px 16px;"
      "}"
      "QPushButton#PrimaryButton:hover { background: #245fd0; }"
      "QPushButton#SecondaryButton, QPushButton#CopyButton, QPushButton#DeleteButton {"
      "  background: #ffffff;"
      "  border: 1px solid #cfd5df;"
      "  border-radius: 7px;"
      "  color: #20242a;"
      "  padding: 6px 12px;"
      "}"
      "QPushButton#SecondaryButton:hover, QPushButton#CopyButton:hover, QPushButton#DeleteButton:hover {"
      "  background: #eef2f7;"
      "}"
      "QWidget#KeybindPill {"
      "  background: #eef2f7;"
      "  border: 1px solid #cfd5df;"
      "  border-radius: 14px;"
      "}"
      "QToolButton#IconButton {"
      "  background: transparent;"
      "  border: none;"
      "  padding: 2px;"
      "}"
      "QToolButton#IconButton:hover { background: #dfe5ef; border-radius: 5px; }"
      "QLabel#KeybindLabel { font-family: 'Menlo'; font-size: 11px; }"
      "QLabel#Title { color: #161a20; font-size: 15px; font-weight: 600; }"
      "QLabel#SectionLabel { color: #3a404a; font-size: 12px; font-weight: 600; }"
      "QLabel#SectionValue, QLabel#StatusLabel { color: #667080; font-size: 11px; }"
      "QLabel#ClipText { color: #242a32; font-size: 12px; }"
      "QDialog#KeybindDialog {"
      "  background: #ffffff;"
      "  border: 1px solid #cfd5df;"
      "}"
      "QScrollArea { background: transparent; border: none; }"
      "QScrollArea > QWidget > QWidget { background: transparent; }";

  qApp->setStyleSheet(style);
}

void MainWindow::loadData() {
  if (!store.load(notes, clips, hotkey, settings)) {
    hotkey = Hotkey::defaultHotkey();
    settings = Settings();
    notes.clear();
    clips.clear();
    store.save(notes, clips, hotkey, settings);
  }
}

void MainWindow::refreshSnippetsUi() {
  clearLayout(clipsListLayout);
  bool consumedHighlight = false;

  if (clips.isEmpty()) {
    auto *emptyLabel = new QLabel("No saved snippets yet", this);
    emptyLabel->setObjectName("SectionValue");
    emptyLabel->setAlignment(Qt::AlignCenter);
    clipsListLayout->addWidget(emptyLabel);
  }

  for (const auto &clip : clips) {
    auto *item = new ClipItemWidget(clip, this);
    connect(item, &ClipItemWidget::copyClicked, this,
            [this](const QString &text) {
              clipboard->setText(text);
              setStatus("Copied");
            });
    connect(item, &ClipItemWidget::deleteClicked, this,
            [this](const QString &text) {
              clips.erase(std::remove_if(clips.begin(), clips.end(),
                                         [&text](const ClipItem &entry) {
                                           return entry.text == text;
                                         }),
                          clips.end());
              refreshSnippetsUi();
              saveData();
            });
    clipsListLayout->addWidget(item);

    if (!consumedHighlight && clip.text == lastAddedClipText) {
      animateFadeIn(item, 180);
      consumedHighlight = true;
    }
  }

  clipsListLayout->addStretch(1);
  clipsCountLabel->setText(QString("%1 saved").arg(clips.size()));
  if (consumedHighlight) {
    lastAddedClipText.clear();
  }
}

void MainWindow::saveSnippetFromInput() {
  const QString text = noteInput->toPlainText().trimmed();
  if (text.isEmpty()) {
    setStatus("Nothing to save");
    return;
  }

  clips.erase(std::remove_if(clips.begin(), clips.end(),
                             [&text](const ClipItem &entry) {
                               return entry.text == text;
                             }),
              clips.end());

  clips.push_front(createClip(text));
  lastAddedClipText = text;
  noteInput->clear();

  refreshSnippetsUi();
  saveData();
}

void MainWindow::toggleVisibility() {
  if (isVisible()) {
    animateHide();
  } else {
    show();
    raise();
    activateWindow();
    noteInput->setFocus();
    animateShow();
  }
}

void MainWindow::beginHotkeyCapture() {
  auto *dialog = new QDialog(this);
  dialog->setObjectName("KeybindDialog");
  dialog->setWindowTitle("Set Shortcut");
  dialog->setModal(true);
  dialog->setFixedSize(340, 160);

  auto *layout = new QVBoxLayout(dialog);
  layout->setContentsMargins(18, 18, 18, 18);
  layout->setSpacing(12);

  auto *title = new QLabel("Press the shortcut to open QuickNote", dialog);
  title->setObjectName("SectionLabel");
  auto *hint = new QLabel("Waiting for input", dialog);
  hint->setObjectName("SectionValue");

  auto *buttonsRow = new QHBoxLayout();
  auto *resetButton = new QPushButton("Use fn + 0", dialog);
  auto *cancelButton = new QPushButton("Cancel", dialog);

  buttonsRow->addWidget(resetButton);
  buttonsRow->addStretch(1);
  buttonsRow->addWidget(cancelButton);

  layout->addWidget(title, 0, Qt::AlignCenter);
  layout->addWidget(hint, 0, Qt::AlignCenter);
  layout->addStretch(1);
  layout->addLayout(buttonsRow);

  connect(cancelButton, &QPushButton::clicked, dialog, &QDialog::reject);
  connect(resetButton, &QPushButton::clicked, this, [this, dialog]() {
    hotkey = Hotkey::defaultHotkey();
    keybindPill->setText(hotkey.toDisplayString());
    hotkeyManager->setHotkey(hotkey);
    saveData();
    dialog->accept();
  });

  connect(dialog, &QDialog::finished, this, [this](int) {
    hotkeyManager->captureNext(nullptr);
  });

  hotkeyManager->captureNext([this, dialog](const Hotkey &captured) {
    hotkey = captured;
    keybindPill->setText(hotkey.toDisplayString());
    hotkeyManager->setHotkey(hotkey);
    saveData();
    dialog->accept();
  });

  dialog->exec();
}

void MainWindow::saveData() {
  store.save(notes, clips, hotkey, settings);
  setStatus("Saved");
}

void MainWindow::setStatus(const QString &text) {
  statusLabel->setText(text);
}

bool MainWindow::eventFilter(QObject *obj, QEvent *event) {
  if (obj == topBar) {
    if (event->type() == QEvent::MouseButtonPress) {
      auto *mouseEvent = static_cast<QMouseEvent *>(event);
      if (mouseEvent->button() == Qt::LeftButton) {
        dragging = true;
        dragStart =
            mouseEvent->globalPosition().toPoint() - frameGeometry().topLeft();
        return true;
      }
    } else if (event->type() == QEvent::MouseMove && dragging) {
      auto *mouseEvent = static_cast<QMouseEvent *>(event);
      move(mouseEvent->globalPosition().toPoint() - dragStart);
      return true;
    } else if (event->type() == QEvent::MouseButtonRelease) {
      dragging = false;
      return true;
    }
  }
  return QMainWindow::eventFilter(obj, event);
}

void MainWindow::focusOutEvent(QFocusEvent *event) {
  QMainWindow::focusOutEvent(event);
  if (settings.autoHide && !QApplication::activeModalWidget()) {
    animateHide();
  }
}

void MainWindow::createTray() {
  if (!QSystemTrayIcon::isSystemTrayAvailable()) {
    return;
  }

  trayMenu = new QMenu(this);
  auto *toggleAction = trayMenu->addAction("Show/Hide QuickNote");
  trayMenu->addSeparator();
  auto *autoHideAction = trayMenu->addAction("Auto-hide on focus loss");
  autoHideAction->setCheckable(true);
  autoHideAction->setChecked(settings.autoHide);
  trayMenu->addSeparator();
  auto *quitAction = trayMenu->addAction("Quit");

  trayIcon = new QSystemTrayIcon(createTrayIcon(), this);
  trayIcon->setToolTip("QuickNote");
  trayIcon->setContextMenu(trayMenu);
  trayIcon->show();

  connect(toggleAction, &QAction::triggered, this,
          &MainWindow::toggleVisibility);
  connect(autoHideAction, &QAction::toggled, this, [this](bool enabled) {
    settings.autoHide = enabled;
    saveData();
  });
  connect(quitAction, &QAction::triggered, qApp, &QApplication::quit);

  connect(trayIcon, &QSystemTrayIcon::activated, this,
          [this](QSystemTrayIcon::ActivationReason reason) {
            if (reason == QSystemTrayIcon::Trigger) {
              toggleVisibility();
            }
          });
}

void MainWindow::animateShow() {
  if (opacityAnimation) {
    opacityAnimation->stop();
  }
  setWindowOpacity(0.0);
  opacityAnimation = new QPropertyAnimation(this, "windowOpacity");
  opacityAnimation->setDuration(160);
  opacityAnimation->setStartValue(0.0);
  opacityAnimation->setEndValue(1.0);
  opacityAnimation->setEasingCurve(QEasingCurve::OutCubic);
  opacityAnimation->start(QAbstractAnimation::DeleteWhenStopped);
}

void MainWindow::animateHide() {
  if (!isVisible()) {
    return;
  }
  if (opacityAnimation) {
    opacityAnimation->stop();
  }
  opacityAnimation = new QPropertyAnimation(this, "windowOpacity");
  opacityAnimation->setDuration(120);
  opacityAnimation->setStartValue(windowOpacity());
  opacityAnimation->setEndValue(0.0);
  opacityAnimation->setEasingCurve(QEasingCurve::InCubic);
  connect(opacityAnimation, &QPropertyAnimation::finished, this, [this]() {
    hide();
    setWindowOpacity(1.0);
  });
  opacityAnimation->start(QAbstractAnimation::DeleteWhenStopped);
}

void MainWindow::animateFadeIn(QWidget *widget, int durationMs) {
  auto *effect = new QGraphicsOpacityEffect(widget);
  widget->setGraphicsEffect(effect);
  auto *animation = new QPropertyAnimation(effect, "opacity", widget);
  animation->setDuration(durationMs);
  animation->setStartValue(0.0);
  animation->setEndValue(1.0);
  animation->setEasingCurve(QEasingCurve::OutCubic);
  animation->start(QAbstractAnimation::DeleteWhenStopped);
}
