#pragma once

#include <QWidget>

#include "../DataStore.h"

class QLabel;
class QPushButton;

class ClipItemWidget : public QWidget {
  Q_OBJECT
public:
  explicit ClipItemWidget(const ClipItem &clip, QWidget *parent = nullptr);
  QString text() const;

signals:
  void copyClicked(const QString &text);
  void deleteClicked(const QString &text);

private:
  const QString clipText;
  QLabel *textLabel = nullptr;
  QPushButton *copyButton = nullptr;
  QPushButton *deleteButton = nullptr;
};
