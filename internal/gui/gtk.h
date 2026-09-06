#ifndef NUVEM_GTK_H
#define NUVEM_GTK_H

void nuvem_run(void);
void nuvem_set_portuguese(int enabled);
void nuvem_set_status(const char *text, int is_error);
void nuvem_set_dashboard(const char *local, const char *remote, int minutes, const char *service, const char *client_id, const char *client_secret, int configured);

#endif
