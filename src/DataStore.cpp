#include "DataStore.h"

#include <QDir>
#include <QFile>
#include <QJsonArray>
#include <QJsonDocument>
#include <QJsonObject>
#include <QStandardPaths>

QString DataStore::dataPath() const {
  QString base =
      QStandardPaths::writableLocation(QStandardPaths::AppDataLocation);
  QDir dir(base);
  if (!dir.exists()) {
    dir.mkpath(".");
  }
  return dir.filePath("stash.json");
}

bool DataStore::load(QVector<Note> &notes, QVector<ClipItem> &clips,
                     Hotkey &hotkey, Settings &settings) {
  QFile file(dataPath());
  if (!file.exists() || !file.open(QIODevice::ReadOnly)) {
    return false;
  }

  QJsonDocument doc = QJsonDocument::fromJson(file.readAll());
  if (!doc.isObject()) {
    return false;
  }

  QJsonObject root = doc.object();
  notes.clear();
  clips.clear();
  settings = Settings();

  for (const auto &value : root.value("notes").toArray()) {
    QJsonObject obj = value.toObject();
    Note note;
    note.id = obj.value("id").toString();
    note.body = obj.value("body").toString();
    note.pinned = obj.value("pinned").toBool(false);
    note.updated =
        QDateTime::fromString(obj.value("updated").toString(), Qt::ISODate);
    if (!note.updated.isValid()) {
      note.updated = QDateTime::currentDateTime();
    }
    notes.push_back(note);
  }

  for (const auto &value : root.value("clips").toArray()) {
    QJsonObject obj = value.toObject();
    ClipItem clip;
    clip.text = obj.value("text").toString();
    clip.captured =
        QDateTime::fromString(obj.value("captured").toString(), Qt::ISODate);
    if (!clip.captured.isValid()) {
      clip.captured = QDateTime::currentDateTime();
    }
    clips.push_back(clip);
  }

  hotkey = Hotkey::fromJson(root.value("hotkey").toObject());
  QJsonObject settingsObj = root.value("settings").toObject();
  settings.autoHide = settingsObj.value("autoHide").toBool(settings.autoHide);
  settings.autoPinTags =
      settingsObj.value("autoPinTags").toBool(settings.autoPinTags);
  return true;
}

bool DataStore::save(const QVector<Note> &notes, const QVector<ClipItem> &clips,
                     const Hotkey &hotkey, const Settings &settings) const {
  QJsonArray notesArray;
  for (const auto &note : notes) {
    QJsonObject obj;
    obj["id"] = note.id;
    obj["body"] = note.body;
    obj["pinned"] = note.pinned;
    obj["updated"] = note.updated.toString(Qt::ISODate);
    notesArray.append(obj);
  }

  QJsonArray clipsArray;
  for (const auto &clip : clips) {
    QJsonObject obj;
    obj["text"] = clip.text;
    obj["captured"] = clip.captured.toString(Qt::ISODate);
    clipsArray.append(obj);
  }

  QJsonObject root;
  root["notes"] = notesArray;
  root["clips"] = clipsArray;
  root["hotkey"] = hotkey.toJson();
  QJsonObject settingsObj;
  settingsObj["autoHide"] = settings.autoHide;
  settingsObj["autoPinTags"] = settings.autoPinTags;
  root["settings"] = settingsObj;

  QFile file(dataPath());
  if (!file.open(QIODevice::WriteOnly | QIODevice::Truncate)) {
    return false;
  }

  file.write(QJsonDocument(root).toJson(QJsonDocument::Indented));
  return true;
}
