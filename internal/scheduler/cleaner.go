package scheduler

import (
	"log"
	"time"

	"g.pervovsky.ru/go-access-admin/internal/access"
	"g.pervovsky.ru/go-access-admin/internal/storage"
)

func StartCleaner(db *storage.DB, interval time.Duration) {
	// Deletes User from .htpasswd file when access is expired id db
	ticker := time.NewTicker(interval)

	go func() {
		for {
			<-ticker.C
			log.Println("[cleaner] checking for expired accesses...")

			expired, err := db.GetExpiredAccesses()
			if err != nil {
				log.Println("[cleaner] error fetching expired: ", err)
				continue
			}

			for _, a := range expired {
				log.Printf("[cleaner] removing user: %s from %s\n", a.Username, a.HtpasswdPath)

				if err := access.RemoveUser(a.HtpasswdPath, a.Username); err != nil {
					log.Printf("[cleaner] error removing from htpasswd: %v\n", err)
					continue
				}

				if err := db.DeleteAccess(int64(a.ID)); err != nil {
					log.Printf("[cleaner] error deleting from db: %v\n", err)
				}
			}

			if len(expired) > 0 {
				log.Printf("[cleaner] removed %d expired accesses\n", len(expired))
			}
		}
	}()
}
