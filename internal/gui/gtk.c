#include <gtk/gtk.h>
#include "gtk.h"

extern void goNuvemActivated(void);
extern void goNuvemSyncNow(void);
extern void goNuvemSaveConfig(char *local, char *remote, int minutes, char *client_id, char *client_secret);
extern void goNuvemConnectGoogle(char *client_id, char *client_secret);

static GtkWidget *window;
static GtkWidget *status_label;
static GtkWidget *folder_label;
static GtkWidget *service_label;
static GtkWidget *local_entry;
static GtkWidget *remote_entry;
static GtkWidget *minutes_spin;
static GtkWidget *sync_button;
static GtkWidget *title_label;
static GtkWidget *description_label;
static GtkWidget *connection_frame;
static GtkWidget *local_label;
static GtkWidget *remote_label;
static GtkWidget *frequency_label;
static GtkWidget *client_id_label;
static GtkWidget *client_secret_label;
static GtkWidget *client_id_entry;
static GtkWidget *client_secret_entry;
static GtkWidget *save_button;
static GtkWidget *choose_button;
static GtkWidget *connect_button;
static int portuguese;

static void set_texts(void) {
  gtk_window_set_title(GTK_WINDOW(window), "Nuvem");
  gtk_label_set_text(GTK_LABEL(title_label), "Nuvem");
  gtk_label_set_text(GTK_LABEL(description_label), portuguese ? "Sua pasta e a nuvem, sem comandos." : "Your folder and the cloud, without commands.");
  gtk_frame_set_label(GTK_FRAME(connection_frame), portuguese ? "Sincronização" : "Synchronization");
  gtk_label_set_text(GTK_LABEL(local_label), portuguese ? "Pasta neste computador" : "Folder on this computer");
  gtk_label_set_text(GTK_LABEL(remote_label), portuguese ? "Pasta remota (somente configuração legada)" : "Remote folder (legacy setup only)");
  gtk_label_set_text(GTK_LABEL(frequency_label), portuguese ? "Verificar a cada (minutos)" : "Check every (minutes)");
  gtk_label_set_text(GTK_LABEL(client_id_label), portuguese ? "ID do cliente OAuth (recomendado)" : "OAuth client ID (recommended)");
  gtk_label_set_text(GTK_LABEL(client_secret_label), portuguese ? "Segredo do cliente OAuth" : "OAuth client secret");
  gtk_button_set_label(GTK_BUTTON(choose_button), portuguese ? "Escolher pasta" : "Choose folder");
  gtk_button_set_label(GTK_BUTTON(save_button), portuguese ? "Salvar configuração" : "Save setup");
  gtk_button_set_label(GTK_BUTTON(connect_button), portuguese ? "Conectar e escolher pasta" : "Connect and choose folder");
  gtk_button_set_label(GTK_BUTTON(sync_button), portuguese ? "Sincronizar agora" : "Sync now");
}

static void on_sync(GtkButton *button, gpointer data) {
  gtk_widget_set_sensitive(GTK_WIDGET(button), FALSE);
  goNuvemSyncNow();
}

static void on_save(GtkButton *button, gpointer data) {
  const char *local = gtk_editable_get_text(GTK_EDITABLE(local_entry));
  const char *remote = gtk_editable_get_text(GTK_EDITABLE(remote_entry));
  const char *client_id = gtk_editable_get_text(GTK_EDITABLE(client_id_entry));
  const char *client_secret = gtk_editable_get_text(GTK_EDITABLE(client_secret_entry));
  int minutes = (int)gtk_spin_button_get_value(GTK_SPIN_BUTTON(minutes_spin));
  goNuvemSaveConfig((char *)local, (char *)remote, minutes, (char *)client_id, (char *)client_secret);
}

static void on_connect(GtkButton *button, gpointer data) {
  const char *client_id = gtk_editable_get_text(GTK_EDITABLE(client_id_entry));
  const char *client_secret = gtk_editable_get_text(GTK_EDITABLE(client_secret_entry));
  gtk_widget_set_sensitive(GTK_WIDGET(button), FALSE);
  goNuvemConnectGoogle((char *)client_id, (char *)client_secret);
}

static void on_folder_selected(GObject *source, GAsyncResult *result, gpointer data) {
  GError *error = NULL;
  GFile *folder = gtk_file_dialog_select_folder_finish(GTK_FILE_DIALOG(source), result, &error);
  if (folder != NULL) {
    char *path = g_file_get_path(folder);
    gtk_editable_set_text(GTK_EDITABLE(local_entry), path);
    g_free(path);
    g_object_unref(folder);
  }
  if (error != NULL) g_error_free(error);
}

static void on_choose_folder(GtkButton *button, gpointer data) {
  GtkFileDialog *dialog = gtk_file_dialog_new();
  gtk_file_dialog_set_title(dialog, portuguese ? "Escolha a pasta local" : "Choose local folder");
  gtk_file_dialog_select_folder(dialog, GTK_WINDOW(window), NULL, on_folder_selected, NULL);
  g_object_unref(dialog);
}

static void activate(GtkApplication *app, gpointer data) {
  window = gtk_application_window_new(app);
  gtk_window_set_default_size(GTK_WINDOW(window), 720, 560);

  GtkWidget *root = gtk_box_new(GTK_ORIENTATION_VERTICAL, 20);
  gtk_widget_set_margin_top(root, 28);
  gtk_widget_set_margin_bottom(root, 28);
  gtk_widget_set_margin_start(root, 32);
  gtk_widget_set_margin_end(root, 32);
  gtk_window_set_child(GTK_WINDOW(window), root);

  title_label = gtk_label_new("Nuvem");
  gtk_widget_set_halign(title_label, GTK_ALIGN_START);
  gtk_widget_add_css_class(title_label, "title-1");
  gtk_box_append(GTK_BOX(root), title_label);

  description_label = gtk_label_new(NULL);
  gtk_widget_set_halign(description_label, GTK_ALIGN_START);
  gtk_widget_add_css_class(description_label, "dim-label");
  gtk_box_append(GTK_BOX(root), description_label);

  status_label = gtk_label_new(NULL);
  gtk_widget_set_halign(status_label, GTK_ALIGN_START);
  gtk_label_set_wrap(GTK_LABEL(status_label), TRUE);
  gtk_box_append(GTK_BOX(root), status_label);

  folder_label = gtk_label_new(NULL);
  gtk_widget_set_halign(folder_label, GTK_ALIGN_START);
  gtk_label_set_wrap(GTK_LABEL(folder_label), TRUE);
  gtk_box_append(GTK_BOX(root), folder_label);

  service_label = gtk_label_new(NULL);
  gtk_widget_set_halign(service_label, GTK_ALIGN_START);
  gtk_widget_add_css_class(service_label, "dim-label");
  gtk_box_append(GTK_BOX(root), service_label);

  connection_frame = gtk_frame_new(NULL);
  GtkWidget *form = gtk_box_new(GTK_ORIENTATION_VERTICAL, 14);
  gtk_widget_set_margin_top(form, 16);
  gtk_widget_set_margin_bottom(form, 16);
  gtk_widget_set_margin_start(form, 16);
  gtk_widget_set_margin_end(form, 16);
  gtk_frame_set_child(GTK_FRAME(connection_frame), form);
  gtk_box_append(GTK_BOX(root), connection_frame);

  local_label = gtk_label_new(NULL);
  local_entry = gtk_entry_new();
  gtk_entry_set_placeholder_text(GTK_ENTRY(local_entry), "/home/name/Documents");
  GtkWidget *local_row = gtk_box_new(GTK_ORIENTATION_HORIZONTAL, 8);
  gtk_widget_set_hexpand(local_entry, TRUE);
  choose_button = gtk_button_new();
  g_signal_connect(choose_button, "clicked", G_CALLBACK(on_choose_folder), NULL);
  gtk_box_append(GTK_BOX(local_row), local_entry);
  gtk_box_append(GTK_BOX(local_row), choose_button);
  gtk_widget_set_halign(local_label, GTK_ALIGN_START);
  gtk_box_append(GTK_BOX(form), local_label);
  gtk_box_append(GTK_BOX(form), local_row);

  remote_label = gtk_label_new(NULL);
  remote_entry = gtk_entry_new();
  gtk_entry_set_placeholder_text(GTK_ENTRY(remote_entry), "GoogleDrive:");
  gtk_widget_set_halign(remote_label, GTK_ALIGN_START);
  gtk_box_append(GTK_BOX(form), remote_label);
  gtk_box_append(GTK_BOX(form), remote_entry);

  frequency_label = gtk_label_new(NULL);
  minutes_spin = gtk_spin_button_new_with_range(1, 1440, 1);
  gtk_spin_button_set_value(GTK_SPIN_BUTTON(minutes_spin), 15);
  gtk_widget_set_halign(frequency_label, GTK_ALIGN_START);
  gtk_box_append(GTK_BOX(form), frequency_label);
  gtk_box_append(GTK_BOX(form), minutes_spin);

  client_id_label = gtk_label_new(NULL);
  gtk_widget_set_halign(client_id_label, GTK_ALIGN_START);
  client_id_entry = gtk_entry_new();
  gtk_entry_set_placeholder_text(GTK_ENTRY(client_id_entry), "1234567890-…apps.googleusercontent.com");
  gtk_box_append(GTK_BOX(form), client_id_label);
  gtk_box_append(GTK_BOX(form), client_id_entry);

  client_secret_label = gtk_label_new(NULL);
  gtk_widget_set_halign(client_secret_label, GTK_ALIGN_START);
  client_secret_entry = gtk_password_entry_new();
  gtk_box_append(GTK_BOX(form), client_secret_label);
  gtk_box_append(GTK_BOX(form), client_secret_entry);

  GtkWidget *actions = gtk_box_new(GTK_ORIENTATION_HORIZONTAL, 10);
  save_button = gtk_button_new();
  g_signal_connect(save_button, "clicked", G_CALLBACK(on_save), NULL);
  connect_button = gtk_button_new();
  g_signal_connect(connect_button, "clicked", G_CALLBACK(on_connect), NULL);
  sync_button = gtk_button_new();
  gtk_widget_add_css_class(sync_button, "suggested-action");
  g_signal_connect(sync_button, "clicked", G_CALLBACK(on_sync), NULL);
  gtk_box_append(GTK_BOX(actions), save_button);
  gtk_box_append(GTK_BOX(actions), connect_button);
  gtk_box_append(GTK_BOX(actions), sync_button);
  gtk_box_append(GTK_BOX(root), actions);

  set_texts();
  gtk_window_present(GTK_WINDOW(window));
  goNuvemActivated();
}

typedef struct { char *text; int is_error; } StatusUpdate;
static gboolean update_status(gpointer data) {
  StatusUpdate *update = data;
  gtk_label_set_text(GTK_LABEL(status_label), update->text);
  if (update->is_error) gtk_widget_add_css_class(status_label, "error");
  else gtk_widget_remove_css_class(status_label, "error");
  g_free(update->text); g_free(update); return G_SOURCE_REMOVE;
}
void nuvem_set_status(const char *text, int is_error) {
  StatusUpdate *update = g_new(StatusUpdate, 1);
  update->text = g_strdup(text); update->is_error = is_error;
  g_idle_add(update_status, update);
}

typedef struct { char *local; char *remote; char *service; char *client_id; char *client_secret; int minutes; int configured; } DashboardUpdate;
static gboolean update_dashboard(gpointer data) {
  DashboardUpdate *update = data;
  gtk_editable_set_text(GTK_EDITABLE(local_entry), update->local);
  gtk_editable_set_text(GTK_EDITABLE(remote_entry), update->remote);
  gtk_editable_set_text(GTK_EDITABLE(client_id_entry), update->client_id);
  gtk_editable_set_text(GTK_EDITABLE(client_secret_entry), update->client_secret);
  gtk_spin_button_set_value(GTK_SPIN_BUTTON(minutes_spin), update->minutes > 0 ? update->minutes : 15);
  gtk_label_set_text(GTK_LABEL(folder_label), update->configured ? update->local : (portuguese ? "Escolha uma pasta local; depois conecte e escolha a pasta do Google Drive." : "Choose a local folder; then connect and choose the Google Drive folder."));
  gtk_label_set_text(GTK_LABEL(service_label), update->service);
  gtk_widget_set_sensitive(sync_button, update->configured);
  gtk_widget_set_sensitive(connect_button, TRUE);
  g_free(update->local); g_free(update->remote); g_free(update->service); g_free(update->client_id); g_free(update->client_secret); g_free(update); return G_SOURCE_REMOVE;
}
void nuvem_set_dashboard(const char *local, const char *remote, int minutes, const char *service, const char *client_id, const char *client_secret, int configured) {
  DashboardUpdate *update = g_new(DashboardUpdate, 1);
  update->local = g_strdup(local); update->remote = g_strdup(remote); update->service = g_strdup(service);
  update->client_id = g_strdup(client_id); update->client_secret = g_strdup(client_secret);
  update->minutes = minutes; update->configured = configured;
  g_idle_add(update_dashboard, update);
}

static gboolean update_language(gpointer data) { if (window != NULL) set_texts(); return G_SOURCE_REMOVE; }
void nuvem_set_portuguese(int enabled) { portuguese = enabled; if (window != NULL) g_idle_add(update_language, NULL); }

void nuvem_run(void) {
  GtkApplication *app = gtk_application_new("io.github.adenauersampaio.Nuvem", G_APPLICATION_DEFAULT_FLAGS);
  g_signal_connect(app, "activate", G_CALLBACK(activate), NULL);
  g_application_run(G_APPLICATION(app), 0, NULL);
  g_object_unref(app);
}
