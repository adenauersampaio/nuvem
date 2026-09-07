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
static GtkWidget *about_button;
static int portuguese;

static char *url_bmc = NULL;
static char *url_livepix = NULL;
static char *url_binance = NULL;
static char *url_github = NULL;

void nuvem_set_donation_urls(const char *bmc, const char *livepix, const char *binance, const char *github) {
  if (url_bmc) g_free(url_bmc);
  if (url_livepix) g_free(url_livepix);
  if (url_binance) g_free(url_binance);
  if (url_github) g_free(url_github);

  url_bmc = (bmc && bmc[0]) ? g_strdup(bmc) : g_strdup("https://buymeacoffee.com/adetech");
  url_livepix = (livepix && livepix[0]) ? g_strdup(livepix) : g_strdup("https://livepix.gg/adetech");
  url_binance = (binance && binance[0]) ? g_strdup(binance) : g_strdup("https://app.binance.com/uni-qr/request-to-pay?billOrderId=452917517181927424&billType=request_a_payment");
  url_github = (github && github[0]) ? g_strdup(github) : g_strdup("https://github.com/sponsors/adenauersampaio");
}

static void on_about(GtkButton *button, gpointer data) {
  GtkWidget *dialog = gtk_window_new();
  gtk_window_set_title(GTK_WINDOW(dialog), portuguese ? "Sobre o Nuvem" : "About Nuvem");
  gtk_window_set_transient_for(GTK_WINDOW(dialog), GTK_WINDOW(window));
  gtk_window_set_modal(GTK_WINDOW(dialog), TRUE);
  gtk_window_set_default_size(GTK_WINDOW(dialog), 480, -1);
  gtk_window_set_resizable(GTK_WINDOW(dialog), FALSE);

  GtkWidget *box = gtk_box_new(GTK_ORIENTATION_VERTICAL, 14);
  gtk_widget_set_margin_top(box, 24);
  gtk_widget_set_margin_bottom(box, 24);
  gtk_widget_set_margin_start(box, 28);
  gtk_widget_set_margin_end(box, 28);
  gtk_window_set_child(GTK_WINDOW(dialog), box);

  GtkWidget *title = gtk_label_new("Nuvem");
  gtk_widget_set_halign(title, GTK_ALIGN_START);
  gtk_widget_add_css_class(title, "title-2");
  gtk_box_append(GTK_BOX(box), title);

  GtkWidget *desc = gtk_label_new(portuguese 
      ? "Sua pasta e a nuvem, sem comandos."
      : "Your folder and the cloud, without commands.");
  gtk_widget_set_halign(desc, GTK_ALIGN_START);
  gtk_widget_add_css_class(desc, "dim-label");
  gtk_box_append(GTK_BOX(box), desc);

  GtkWidget *prompt = gtk_label_new(portuguese
      ? "☕ Se este software está sendo útil, cogite deixar um café... US$ 1"
      : "☕ If this software is useful to you, consider buying me a coffee... $1");
  gtk_widget_set_halign(prompt, GTK_ALIGN_START);
  gtk_label_set_wrap(GTK_LABEL(prompt), TRUE);
  gtk_box_append(GTK_BOX(box), prompt);

  const char *bmc = (url_bmc && url_bmc[0]) ? url_bmc : "https://buymeacoffee.com/adetech";
  const char *livepix = (url_livepix && url_livepix[0]) ? url_livepix : "https://livepix.gg/adetech";
  const char *binance = (url_binance && url_binance[0]) ? url_binance : "https://app.binance.com/uni-qr/request-to-pay?billOrderId=452917517181927424&billType=request_a_payment";
  const char *github = (url_github && url_github[0]) ? url_github : "https://github.com/sponsors/adenauersampaio";

  GtkWidget *links_box = gtk_box_new(GTK_ORIENTATION_VERTICAL, 8);
  gtk_widget_set_margin_top(links_box, 6);
  gtk_widget_set_margin_bottom(links_box, 6);

  GtkWidget *btn_livepix = gtk_link_button_new_with_label(
      livepix,
      portuguese ? "🇧🇷 Apoiar via Pix (Livepix)" : "🇧🇷 Support with Pix (Livepix)");
  GtkWidget *btn_bmc = gtk_link_button_new_with_label(
      bmc,
      portuguese ? "☕ Deixar um café (Buy Me a Coffee)" : "☕ Buy Me a Coffee ($1)");
  GtkWidget *btn_binance = gtk_link_button_new_with_label(
      binance,
      portuguese ? "💛 Apoiar via Binance Pay (Cripto)" : "💛 Support via Binance Pay (Crypto)");
  GtkWidget *btn_github = gtk_link_button_new_with_label(
      github,
      "💖 GitHub Sponsors");

  if (portuguese) {
    gtk_box_append(GTK_BOX(links_box), btn_livepix);
    gtk_box_append(GTK_BOX(links_box), btn_bmc);
    gtk_box_append(GTK_BOX(links_box), btn_binance);
    gtk_box_append(GTK_BOX(links_box), btn_github);
  } else {
    gtk_box_append(GTK_BOX(links_box), btn_bmc);
    gtk_box_append(GTK_BOX(links_box), btn_github);
    gtk_box_append(GTK_BOX(links_box), btn_binance);
    gtk_box_append(GTK_BOX(links_box), btn_livepix);
  }
  gtk_box_append(GTK_BOX(box), links_box);

  GtkWidget *close_btn = gtk_button_new_with_label(portuguese ? "Fechar" : "Close");
  gtk_widget_set_halign(close_btn, GTK_ALIGN_END);
  g_signal_connect_swapped(close_btn, "clicked", G_CALLBACK(gtk_window_destroy), dialog);
  gtk_box_append(GTK_BOX(box), close_btn);

  gtk_window_present(GTK_WINDOW(dialog));
}

static void set_texts(void) {
  gtk_window_set_title(GTK_WINDOW(window), "Nuvem");
  gtk_label_set_text(GTK_LABEL(title_label), "Nuvem");
  if (about_button != NULL) {
    gtk_button_set_label(GTK_BUTTON(about_button), portuguese ? "Sobre" : "About");
    gtk_widget_set_tooltip_text(about_button, portuguese ? "Sobre o Nuvem e apoio" : "About Nuvem and support");
  }
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
  gtk_window_set_default_size(GTK_WINDOW(window), 680, 520);

  GtkWidget *scrolled = gtk_scrolled_window_new();
  gtk_scrolled_window_set_policy(GTK_SCROLLED_WINDOW(scrolled), GTK_POLICY_NEVER, GTK_POLICY_AUTOMATIC);
  gtk_scrolled_window_set_propagate_natural_height(GTK_SCROLLED_WINDOW(scrolled), TRUE);
  gtk_window_set_child(GTK_WINDOW(window), scrolled);

  GtkWidget *root = gtk_box_new(GTK_ORIENTATION_VERTICAL, 10);
  gtk_widget_set_margin_top(root, 14);
  gtk_widget_set_margin_bottom(root, 14);
  gtk_widget_set_margin_start(root, 20);
  gtk_widget_set_margin_end(root, 20);
  gtk_scrolled_window_set_child(GTK_SCROLLED_WINDOW(scrolled), root);

  GtkWidget *header_row = gtk_box_new(GTK_ORIENTATION_HORIZONTAL, 12);
  title_label = gtk_label_new("Nuvem");
  gtk_widget_set_halign(title_label, GTK_ALIGN_START);
  gtk_widget_add_css_class(title_label, "title-1");
  gtk_box_append(GTK_BOX(header_row), title_label);

  GtkWidget *spacer = gtk_box_new(GTK_ORIENTATION_HORIZONTAL, 0);
  gtk_widget_set_hexpand(spacer, TRUE);
  gtk_box_append(GTK_BOX(header_row), spacer);

  about_button = gtk_button_new();
  gtk_widget_add_css_class(about_button, "flat");
  g_signal_connect(about_button, "clicked", G_CALLBACK(on_about), NULL);
  gtk_box_append(GTK_BOX(header_row), about_button);

  gtk_box_append(GTK_BOX(root), header_row);

  GtkWidget *info_box = gtk_box_new(GTK_ORIENTATION_VERTICAL, 3);
  description_label = gtk_label_new(NULL);
  gtk_widget_set_halign(description_label, GTK_ALIGN_START);
  gtk_widget_add_css_class(description_label, "dim-label");
  gtk_box_append(GTK_BOX(info_box), description_label);

  status_label = gtk_label_new(NULL);
  gtk_widget_set_halign(status_label, GTK_ALIGN_START);
  gtk_label_set_wrap(GTK_LABEL(status_label), TRUE);
  gtk_box_append(GTK_BOX(info_box), status_label);

  folder_label = gtk_label_new(NULL);
  gtk_widget_set_halign(folder_label, GTK_ALIGN_START);
  gtk_label_set_wrap(GTK_LABEL(folder_label), TRUE);
  gtk_box_append(GTK_BOX(info_box), folder_label);

  service_label = gtk_label_new(NULL);
  gtk_widget_set_halign(service_label, GTK_ALIGN_START);
  gtk_widget_add_css_class(service_label, "dim-label");
  gtk_box_append(GTK_BOX(info_box), service_label);

  gtk_box_append(GTK_BOX(root), info_box);

  connection_frame = gtk_frame_new(NULL);
  GtkWidget *form = gtk_box_new(GTK_ORIENTATION_VERTICAL, 8);
  gtk_widget_set_margin_top(form, 10);
  gtk_widget_set_margin_bottom(form, 10);
  gtk_widget_set_margin_start(form, 12);
  gtk_widget_set_margin_end(form, 12);
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
  GtkWidget *local_group = gtk_box_new(GTK_ORIENTATION_VERTICAL, 3);
  gtk_box_append(GTK_BOX(local_group), local_label);
  gtk_box_append(GTK_BOX(local_group), local_row);
  gtk_box_append(GTK_BOX(form), local_group);

  remote_label = gtk_label_new(NULL);
  remote_entry = gtk_entry_new();
  gtk_entry_set_placeholder_text(GTK_ENTRY(remote_entry), "GoogleDrive:");
  gtk_widget_set_halign(remote_label, GTK_ALIGN_START);
  GtkWidget *remote_group = gtk_box_new(GTK_ORIENTATION_VERTICAL, 3);
  gtk_box_append(GTK_BOX(remote_group), remote_label);
  gtk_box_append(GTK_BOX(remote_group), remote_entry);
  gtk_box_append(GTK_BOX(form), remote_group);

  frequency_label = gtk_label_new(NULL);
  minutes_spin = gtk_spin_button_new_with_range(1, 1440, 1);
  gtk_spin_button_set_value(GTK_SPIN_BUTTON(minutes_spin), 15);
  gtk_widget_set_halign(frequency_label, GTK_ALIGN_START);
  GtkWidget *freq_group = gtk_box_new(GTK_ORIENTATION_VERTICAL, 3);
  gtk_box_append(GTK_BOX(freq_group), frequency_label);
  gtk_box_append(GTK_BOX(freq_group), minutes_spin);
  gtk_box_append(GTK_BOX(form), freq_group);

  client_id_label = gtk_label_new(NULL);
  gtk_widget_set_halign(client_id_label, GTK_ALIGN_START);
  client_id_entry = gtk_entry_new();
  gtk_entry_set_placeholder_text(GTK_ENTRY(client_id_entry), "1234567890-…apps.googleusercontent.com");
  GtkWidget *cid_group = gtk_box_new(GTK_ORIENTATION_VERTICAL, 3);
  gtk_box_append(GTK_BOX(cid_group), client_id_label);
  gtk_box_append(GTK_BOX(cid_group), client_id_entry);
  gtk_box_append(GTK_BOX(form), cid_group);

  client_secret_label = gtk_label_new(NULL);
  gtk_widget_set_halign(client_secret_label, GTK_ALIGN_START);
  client_secret_entry = gtk_password_entry_new();
  GtkWidget *csec_group = gtk_box_new(GTK_ORIENTATION_VERTICAL, 3);
  gtk_box_append(GTK_BOX(csec_group), client_secret_label);
  gtk_box_append(GTK_BOX(csec_group), client_secret_entry);
  gtk_box_append(GTK_BOX(form), csec_group);

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
