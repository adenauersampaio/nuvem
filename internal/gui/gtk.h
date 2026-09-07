#ifndef NUVEM_GTK_H
#define NUVEM_GTK_H

void nuvem_run(void);
void nuvem_set_portuguese(int enabled);
void nuvem_set_status(const char *text, int is_error);
void nuvem_set_dashboard(const char *local, const char *remote, int minutes, const char *service, int configured);
void nuvem_set_donation_urls(const char *bmc, const char *livepix, const char *binance, const char *github);
void nuvem_show_window(void);
void nuvem_quit(void);

#endif
