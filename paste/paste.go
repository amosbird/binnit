/*
 *  This program is free software: you can redistribute it and/or
 *  modify it under the terms of the GNU Affero General Public License as
 *  published by the Free Software Foundation, either version 3 of the
 *  License, or (at your option) any later version.
 *
 *  This program is distributed in the hope that it will be useful,
 *  but WITHOUT ANY WARRANTY; without even the implied warranty of
 *  MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the GNU
 *  General Public License for more details.
 *
 *  You should have received a copy of the GNU Affero General Public
 *  License along with this program.  If not, see
 *  <http://www.gnu.org/licenses/>.
 *
 *  (c) Vincenzo "KatolaZ" Nicosia 2017 -- <katolaz@freaknet.org>
 *
 *
 *  This file is part of "binnit", a minimal no-fuss pastebin-like
 *  server written in golang
 *
 */

/*
 *
 * Store/Retrieve functions for FS-based paste storage
 *
 */

package paste

import (
	"crypto/sha256"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"os"
)

func Store(content, dest_dir string) (string, error) {
	h := sha256.New()
	h.Write([]byte(content))
	paste_hash := fmt.Sprintf("%x", h.Sum(nil))
	log.Printf("  `-- hash: %s\n", paste_hash)
	paste_dir := dest_dir + "/"
	// Now we save the file
	paste_name := paste_hash[0:16]
	if _, err := os.Stat(paste_dir + paste_name); os.IsNotExist(err) {
		// The file does not exist, so we can create it
		if err := ioutil.WriteFile(paste_dir+paste_name, []byte(content), 0644); err == nil {
			// and then we return the URL:
			log.Printf("  `-- saving new paste to : %s", paste_dir+paste_name)
		} else {
			log.Printf("Cannot create the paste: %s!\n", paste_dir+paste_name)
		}
	}
	return paste_name, nil
}

func Retrieve(URI string) (content []byte, err error) {
	content, err = ioutil.ReadFile(URI)
	if err != nil {
		return nil, errors.New("Cannot retrieve paste!!!")
	}
	return content, nil
}
