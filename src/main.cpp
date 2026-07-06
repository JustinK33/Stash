#include <QApplication>

#include "MainWindow.h"

int main(int argc, char *argv[]) {
  QApplication app(argc, argv);
  app.setApplicationName("QuickNote");
  app.setOrganizationName("QuickNote");
  app.setOrganizationDomain("quicknote.local");
  app.setQuitOnLastWindowClosed(false);

  MainWindow window;
  window.show();

  return app.exec();
}
