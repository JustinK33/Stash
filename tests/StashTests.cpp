#include <QtTest/QtTest>

#include <QApplication>
#include <QClipboard>
#include <QFile>
#include <QPushButton>
#include <QStandardPaths>
#include <QTextEdit>

#include "DataStore.h"
#include "MainWindow.h"
#include "widgets/ClipItemWidget.h"

class StashTests : public QObject {
  Q_OBJECT

private slots:
  void initTestCase();
  void datastorePersistsSnippetsAndShortcut();
  void uiMigratesLegacyFnZeroShortcut();
  void uiSavesCopiesDeletesAndClearsSnippets();
  void uiCanToggleVisibilityRepeatedly();
  void uiHotkeyToggleRestoresMinimizedWindow();

private:
  QPushButton *buttonByText(QWidget &window, const QString &text);
};

void StashTests::initTestCase() {
  QStandardPaths::setTestModeEnabled(true);
  QCoreApplication::setApplicationName("StashTests");
  QCoreApplication::setOrganizationName("StashTests");
  QCoreApplication::setOrganizationDomain("stash.test");

  DataStore store;
  QFile::remove(store.dataPath());
}

void StashTests::datastorePersistsSnippetsAndShortcut() {
  DataStore store;
  QFile::remove(store.dataPath());

  QVector<Note> notes;
  QVector<ClipItem> clips;
  ClipItem clip;
  clip.text = "Saved text";
  clip.captured = QDateTime::currentDateTime();
  clips.push_back(clip);

  Hotkey hotkey = Hotkey::defaultHotkey();
  Settings settings;
  settings.autoHide = false;

  QVERIFY(store.save(notes, clips, hotkey, settings));

  QVector<Note> loadedNotes;
  QVector<ClipItem> loadedClips;
  Hotkey loadedHotkey;
  Settings loadedSettings;

  QVERIFY(store.load(loadedNotes, loadedClips, loadedHotkey, loadedSettings));
  QCOMPARE(loadedClips.size(), 1);
  QCOMPARE(loadedClips.first().text, QString("Saved text"));
  QVERIFY(!loadedHotkey.fn);
  QVERIFY(loadedHotkey.ctrl);
  QVERIFY(loadedHotkey.alt);
  QVERIFY(loadedHotkey.toDisplayString().contains("0"));
  QVERIFY(!loadedSettings.autoHide);
}

void StashTests::uiMigratesLegacyFnZeroShortcut() {
  DataStore store;
  QFile::remove(store.dataPath());

  QVector<Note> notes;
  QVector<ClipItem> clips;
  Hotkey legacyHotkey;
  legacyHotkey.keyCode = 29;
  legacyHotkey.fn = true;
  legacyHotkey.ctrl = false;
  legacyHotkey.alt = false;
  legacyHotkey.shift = false;
  legacyHotkey.cmd = false;
  Settings settings;

  QVERIFY(store.save(notes, clips, legacyHotkey, settings));

  MainWindow window;
  Q_UNUSED(window);

  QVector<Note> loadedNotes;
  QVector<ClipItem> loadedClips;
  Hotkey loadedHotkey;
  Settings loadedSettings;
  QVERIFY(store.load(loadedNotes, loadedClips, loadedHotkey, loadedSettings));
  QVERIFY(!loadedHotkey.fn);
  QVERIFY(loadedHotkey.ctrl);
  QVERIFY(loadedHotkey.alt);
}

void StashTests::uiSavesCopiesDeletesAndClearsSnippets() {
  DataStore store;
  QFile::remove(store.dataPath());

  MainWindow window;
  window.show();
  QVERIFY(QTest::qWaitForWindowExposed(&window));

  auto *input = window.findChild<QTextEdit *>("SnippetInput");
  QVERIFY(input);

  auto *saveButton = buttonByText(window, "Save");
  auto *clearButton = buttonByText(window, "Clear");
  auto *clearAllButton = buttonByText(window, "Clear All");
  QVERIFY(saveButton);
  QVERIFY(clearButton);
  QVERIFY(clearAllButton);

  QTest::mouseClick(clearAllButton, Qt::LeftButton);
  QTest::keyClicks(input, "first saved snippet");
  QTest::mouseClick(saveButton, Qt::LeftButton);
  QCoreApplication::processEvents();

  auto items = window.findChildren<ClipItemWidget *>();
  QCOMPARE(items.size(), 1);
  QCOMPARE(items.first()->text(), QString("first saved snippet"));

  auto *copyButton = buttonByText(window, "Copy");
  QVERIFY(copyButton);
  QTest::mouseClick(copyButton, Qt::LeftButton);
  QCOMPARE(QApplication::clipboard()->text(), QString("first saved snippet"));

  QTest::keyClicks(input, "temporary text");
  QTest::mouseClick(clearButton, Qt::LeftButton);
  QCOMPARE(input->toPlainText(), QString());

  QTest::keyClicks(input, "second saved snippet");
  QTest::mouseClick(saveButton, Qt::LeftButton);
  QCoreApplication::processEvents();
  QCOMPARE(window.findChildren<ClipItemWidget *>().size(), 2);

  auto *deleteButton = buttonByText(window, "Delete");
  QVERIFY(deleteButton);
  QTest::mouseClick(deleteButton, Qt::LeftButton);
  QCoreApplication::processEvents();
  QCOMPARE(window.findChildren<ClipItemWidget *>().size(), 1);

  QTest::mouseClick(clearAllButton, Qt::LeftButton);
  QCoreApplication::processEvents();
  QCOMPARE(window.findChildren<ClipItemWidget *>().size(), 0);
}

void StashTests::uiCanToggleVisibilityRepeatedly() {
  DataStore store;
  QFile::remove(store.dataPath());

  MainWindow window;
  window.show();
  QVERIFY(QTest::qWaitForWindowExposed(&window));

  QMetaObject::invokeMethod(&window, "toggleVisibility");
  QTest::qWait(180);
  QVERIFY(!window.isVisible());

  QMetaObject::invokeMethod(&window, "toggleVisibility");
  QTest::qWait(220);
  QVERIFY(window.isVisible());

  QMetaObject::invokeMethod(&window, "toggleVisibility");
  QTest::qWait(180);
  QVERIFY(!window.isVisible());

  QMetaObject::invokeMethod(&window, "toggleVisibility");
  QTest::qWait(220);
  QVERIFY(window.isVisible());
}

void StashTests::uiHotkeyToggleRestoresMinimizedWindow() {
  DataStore store;
  QFile::remove(store.dataPath());

  MainWindow window;
  window.show();
  QVERIFY(QTest::qWaitForWindowExposed(&window));

  window.showMinimized();
  QCoreApplication::processEvents();
  QVERIFY(window.isMinimized());

  QMetaObject::invokeMethod(&window, "toggleVisibility");
  QTest::qWait(220);
  QVERIFY(window.isVisible());
  QVERIFY(!window.isMinimized());
}

QPushButton *StashTests::buttonByText(QWidget &window, const QString &text) {
  const auto buttons = window.findChildren<QPushButton *>();
  for (auto *button : buttons) {
    if (button->text() == text) {
      return button;
    }
  }
  return nullptr;
}

QTEST_MAIN(StashTests)
#include "StashTests.moc"
