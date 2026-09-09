# Code Analysis: cockpit-samba-ad-dc Plugin (GSoC 2020)

**Repository:** https://gitlab.com/samba-team/cockpit-samba-ad-dc (canonical)
**Mirror:** https://github.com/evopsbr/cockpit-samba-ad-dc
**Author:** Hezekiah Maina (University of Nairobi) — GSoC 2020 student
**Mentor:** Alexander Bokovoy (Samba Team, Red Hat)
**License:** LGPL-2.1
**Last commit:** 2020-08-28 (290 commits total)
**SambaXP talk:** https://sambaxp.org/fileadmin/user_upload/sambaxp2021-slides/Bokovoy_cockpit_ui_samba_ad_dc.pdf

---

## 1. Plugin Structure

**Pure JavaScript/React frontend — no Python, no C, no backend daemon.** The plugin
is built on Cockpit's Starter Kit and communicates with the server exclusively through
`cockpit.script()` and `cockpit.spawn()` API calls that shell out to `samba-tool`.

### Technology Stack
| Layer | Technology |
|---|---|
| UI Framework | React (functional components + hooks) |
| Component Library | PatternFly 4 (`@patternfly/react-core`, `@patternfly/react-icons`) |
| Build System | Webpack + npm (`make` wraps `npm run build`) |
| Server Bridge | Cockpit bridge (`cockpit.script`, `cockpit.spawn`) with `superuser: true` |
| Packaging | RPM (.spec.in), DEB (debian.control/rules), OBS automated builds |
| CI | Travis CI (`.travis.yml`) |
| Linting | ESLint (`.eslintrc.json`) |

### File Organization (145 source files in `src/`)

```
src/
├── index.html              # Main entry HTML (loads cockpit.js + index.js)
├── index.js                # React root mount
├── app.jsx                 # Root component → <ServerRole/>
├── ad-dc-status.js         # Gate: checks if server is AD DC via testparm
├── provision-modal.js      # Domain provisioning wizard
├── main.js                 # Dashboard with cards linking to each module
├── common.js               # Shared React components (RenderError, Success, Loading, BackButton, ErrorToast, SuccessToast)
├── manifest.json           # Cockpit manifest (name, version, tools)
├── css/common.css          # Shared CSS
├── lib/                    # PatternFly SCSS overrides
├── computer/               # Computer management (5 operations)
├── contact/                # Contact management (5 operations)
├── delegation/             # Delegation management (5 operations)
├── dns/                    # DNS management (7 operations)
├── domain/                 # Domain management (info, join, demote, dcpromo, classicupgrade, backup/*, trust/*)
│   ├── backup/             # 4 backup operations
│   └── trust/              # 6 trust operations
├── dsacl/                  # DS ACLs manipulation (get, set)
├── forest/                 # Forest management (show, dsheuristics)
├── fsmo/                   # FSMO management (show, seize, transfer)
├── gpo/                    # GPO management (14 operations)
├── group/                  # Group management (7 operations)
├── ntacl/                  # NT ACLs manipulation (6 operations)
├── organization_unit/      # OU management (6 operations)
├── sites/                  # Sites management (5 operations)
├── spn/                    # SPN management (3 operations)
├── time/                   # Server time
└── user/                   # User management (9 operations)
```

Each module directory contains:
- `<module>.html` — Static HTML entry point (loaded by Cockpit as separate page)
- `index.js` — Module React root + page layout
- `<operation>.js` — One file per samba-tool subcommand
- Optional `index.css` — Module-specific styles

---

## 2. EXACT samba-tool Subcommand Coverage

Every `samba-tool` invocation extracted from source code (101 total calls: 91 `cockpit.script` + 10 `cockpit.spawn`):

### Computer Management (`samba-tool computer`)
| File | Command |
|---|---|
| `computer/create.js` | `samba-tool computer create ${computerName}` |
| `computer/delete.js` | `samba-tool computer delete ${computerName}` |
| `computer/list.js` | `samba-tool computer list` |
| `computer/move.js` | `samba-tool computer move ${computerName} ${newOrgUnit}` |
| `computer/show.js` | `samba-tool computer show ${computerName}` |

### Contact Management (`samba-tool contact`)
| File | Command |
|---|---|
| `contact/create.js` | `samba-tool contact create --given-name=${givenName} --initials=${initials} --surname=${surname}` |
| `contact/delete.js` | `samba-tool contact delete ${contactName}` |
| `contact/list.js` | `samba-tool contact list` |
| `contact/move.js` | `samba-tool contact move ${contactName} ${newOrgUnit}` |
| `contact/show.js` | `samba-tool contact show ${contactName}` |

### Delegation Management (`samba-tool delegation`)
| File | Command |
|---|---|
| `delegation/add-service.js` | `samba-tool delegation add-service ${accountName} ${principal}` |
| `delegation/delete-service.js` | `samba-tool delegation del-service ${accountName} ${principal}` |
| `delegation/any-protocol.js` | `samba-tool delegation for-any-protocol ${accountName} ${protocolState}` |
| `delegation/any-service.js` | `samba-tool delegation for-any-service ${accountName} ${serviceState}` |
| `delegation/show.js` | `samba-tool delegation show ${accountName}` |

### DNS Management (`samba-tool dns`)
| File | Command |
|---|---|
| `dns/create.js` | `samba-tool dns add ${server} ${zone} ${name} ${type} ${data} --password=${password}` |
| `dns/delete.js` | `samba-tool dns delete ${server} ${zone} ${name} ${type} ${data} --password=${password}` |
| `dns/cleanup.js` | `samba-tool dns cleanup ${server} ${name} --password=${password}` |
| `dns/serverinfo.js` | `samba-tool dns serverinfo ${server} --password=${password}` |
| `dns/zonecreate.js` | `samba-tool dns zonecreate ${server} ${zone} --password=${password}` |
| `dns/zonedelete.js` | `samba-tool dns zonedelete ${server} ${zone} --password=${password}` |
| `dns/zoneinfo.js` | `samba-tool dns zoneinfo ${server} ${zone} --password=${password}` |
| `dns/zonelist.js` | `samba-tool dns zonelist ${server} --password=${password}` |

### Domain Management (`samba-tool domain`)
| File | Command |
|---|---|
| `domain/info.js` | `samba-tool domain info ${ipAddress}` |
| `domain/dcpromo.js` | `samba-tool domain dcpromo ${dnsDomain} ${role}` |
| `domain/demote.js` | `samba-tool domain demote` (with `--server=`, `--URL=`, `--remove-other-dead-server=` flags) |
| `domain/join.js` | `samba-tool domain join` (with role, domain, adminpass, server, site, dns-backend, etc.) |
| `domain/classicupgrade.js` | `samba-tool domain classicupgrade --dbdir=${dbdir} ${smbconf}` |
| `domain/backup/offline.js` | `samba-tool domain backup offline --targetdir=${targetDir}` |
| `domain/backup/online.js` | `samba-tool backup online --server=${server} --targetdir=${targetDir}` |
| `domain/backup/rename.js` | `samba-tool backup rename ${newDomain} ${newRealm} --server=${server} --targetdir=${targetDir}` |
| `domain/backup/restore.js` | `samba-tool domain backup restore --backup-file=${tarFile} --targetdir=${outputDir} --newservername=${serverName}` |

### Trust Management (`samba-tool domain trust`)
| File | Command |
|---|---|
| `domain/trust/create.js` | `samba-tool domain trust create ${domain}` |
| `domain/trust/delete.js` | `samba-tool domain trust delete ${domain}` |
| `domain/trust/list.js` | `samba-tool domain trust list` |
| `domain/trust/show.js` | `samba-tool domain trust show ${name}` |
| `domain/trust/validate.js` | `samba-tool domain trust validate ${domain}` |
| `domain/trust/namespaces.js` | `samba-tool domain namespaces online ${domain}` |

### Forest Management (`samba-tool forest`)
| File | Command |
|---|---|
| `forest/show.js` | `samba-tool forest directory_service show` |
| `forest/dsheuristics.js` | `samba-tool forest directory_service dsheuristics ${heuristicsValue}` |

### FSMO Management (`samba-tool fsmo`)
| File | Command |
|---|---|
| `fsmo/show.js` | `samba-tool fsmo show` |
| `fsmo/seize.js` | `samba-tool fsmo seize` (with role flags: `--role=...`, admin credentials) |
| `fsmo/transfer.js` | `samba-tool fsmo transfer` (with role flags: `--role=...`, admin credentials) |

### GPO Management (`samba-tool gpo`)
| File | Command |
|---|---|
| `gpo/create.js` | `samba-tool gpo create ${displayName}` |
| `gpo/delete.js` | `samba-tool gpo del ${gpo}` |
| `gpo/backup.js` | `samba-tool gpo backup ${gpo}` |
| `gpo/dellink.js` | `samba-tool gpo dellink ${container} ${gpo}` |
| `gpo/fetch.js` | `samba-tool gpo fetch ${gpo}` |
| `gpo/getinheritance.js` | `samba-tool gpo getinheritance ${container}` |
| `gpo/getlink.js` | `samba-tool gpo getlink ${container}` |
| `gpo/list.js` | `samba-tool gpo list ${account}` |
| `gpo/listall.js` | `samba-tool gpo listall` |
| `gpo/listcontainers.js` | `samba-tool gpo listcontainers ${gpo}` |
| `gpo/restore.js` | `samba-tool gpo restore ${displayName} ${location}` |
| `gpo/setinheritance.js` | `samba-tool gpo setinheritance ${container} ${choice}` |
| `gpo/setlink.js` | `samba-tool gpo setlink ${container} ${gpo}` |
| `gpo/show.js` | `samba-tool gpo show ${gpo}` |

### Group Management (`samba-tool group`)
| File | Command |
|---|---|
| `group/create.js` | `samba-tool group add ${groupName}` |
| `group/delete.js` | `samba-tool group delete ${groupName}` |
| `group/listgroups.js` | `samba-tool group list` |
| `group/listmembers.js` | `samba-tool group listmembers ${groupName}` |
| `group/move.js` | `samba-tool group move ${group} ${ouContainer}` |
| `group/removemembers.js` | `samba-tool group removemembers ${group} ${members}` |
| `group/show.js` | `samba-tool group show ${groupName}` |

### DS ACLs (`samba-tool dsacl`)
| File | Command |
|---|---|
| `dsacl/get.js` | `samba-tool dsacl get` |
| `dsacl/set.js` | `samba-tool dsacl set` (with `--URL=`, `--car=`, `--action=`, `--objectdn=`, `--trusteedn=`, `--sddl=` flags via `cockpit.spawn`) |

### NT ACLs (`samba-tool ntacl`)
| File | Command |
|---|---|
| `ntacl/get.js` | `samba-tool ntacl get ${file}` (with `--xattr-backend=`, `--eadb-file=`, `--use-ntvfs=`, `--use-s3fs=`, `--service=` via `cockpit.spawn`) |
| `ntacl/set.js` | `samba-tool ntacl set ${acl} ${file}` (same optional flags via `cockpit.spawn`) |
| `ntacl/changedomsid.js` | `samba-tool ntacl changedomsid ${origSid} ${newSid} ${file}` (same optional flags via `cockpit.spawn`) |
| `ntacl/getdosinfo.js` | `samba-tool ntacl getdosinfo ${file}` (via `cockpit.spawn`) |
| `ntacl/sysvolcheck.js` | `samba-tool ntacl sysvolcheck ${file}` (via `cockpit.spawn`) |
| `ntacl/sysvolreset.js` | `samba-tool ntacl sysvolreset ${file}` (via `cockpit.spawn`) |

### OU Management (`samba-tool ou`)
| File | Command |
|---|---|
| `organization_unit/create.js` | `samba-tool ou create ${oudn}` (optional `--description=`) |
| `organization_unit/delete.js` | `samba-tool ou delete ${oudn}` |
| `organization_unit/list.js` | `samba-tool ou list` |
| `organization_unit/listobjects.js` | `samba-tool ou listobjects ${oudn}` |
| `organization_unit/move.js` | `samba-tool ou move ${oldOudn} ${newOudn}` |
| `organization_unit/rename.js` | `samba-tool ou rename ${oldOudnName} ${newOudnName}` |

### Sites Management (`samba-tool sites`)
| File | Command |
|---|---|
| `sites/create.js` | `samba-tool sites create ${siteName}` |
| `sites/remove.js` | `samba-tool sites remove ${siteName}` |
| `sites/create_subnet.js` | `samba-tool sites subnet create ${subnet} ${siteOfSubnet}` |
| `sites/remove_subnet.js` | `samba-tool sites subnet remove ${subnet}` |
| `sites/set-site.js` | `samba-tool sites subnet set-site ${subnet} ${siteOfSubnet}` |

### SPN Management (`samba-tool spn`)
| File | Command |
|---|---|
| `spn/add.js` | `samba-tool spn add ${name} ${user}` |
| `spn/delete.js` | `samba-tool spn delete ${name} ${user}` |
| `spn/list.js` | `samba-tool spn list ${user}` |

### User Management (`samba-tool user`)
| File | Command |
|---|---|
| `user/create.js` | `samba-tool user create` (with multiple flags) |
| `user/delete.js` | `samba-tool user delete ${userName}` |
| `user/disable.js` | `samba-tool user disable ${userName}` |
| `user/enable.js` | `samba-tool user enable ${userName}` |
| `user/list.js` | `samba-tool user list` |
| `user/move.js` | `samba-tool user move ${userName} ${orgUnitContainer}` |
| `user/password.js` | `samba-tool user password --password=${password} --newpassword=${newPassword}` |
| `user/setexpiry.js` | `samba-tool user setexpiry ${userName} ${days}` or `samba-tool user setexpiry ${userName} --noexpiry` |
| `user/setpassword.js` | `samba-tool user ${userName} ${password}` (with optional `--must-change-next-login`) |
| `user/show.js` | `samba-tool user show ${userName}` |

### Time
| File | Command |
|---|---|
| `time/index.js` | `samba-tool time ${server}` |

### Provisioning (not a samba-tool subcommand per se, but multi-line script)
| File | Command |
|---|---|
| `provision-modal.js` | `samba-tool domain provision --use-rfc2307 --realm=${realm} --domain=${domain} --server-role=${serverRole} --dns-backend=${dnsBackend} --adminpass=${password1}` followed by `cp -f /var/lib/samba/private/krb5.conf /etc/krb5.conf` then `samba` then `samba-tool domain info 127.0.0.1` |

### Server Role Detection
| File | Command |
|---|---|
| `ad-dc-status.js` | `samba-tool testparm --parameter-name=serverrole` (note: actual binary is `testparm`, invoked via `samba-tool testparm`) |

### Summary: samba-tool Subcommand Groups Covered

| Subcommand Group | Operations Count |
|---|---|
| `computer` | 5 |
| `contact` | 5 |
| `delegation` | 5 |
| `dns` | 8 |
| `domain` (excl. trust) | 9 |
| `domain trust` | 6 |
| `forest` | 2 |
| `fsmo` | 3 |
| `gpo` | 14 |
| `group` | 7 |
| `dsacl` | 2 |
| `ntacl` | 6 |
| `ou` | 6 |
| `sites` | 5 |
| `spn` | 3 |
| `user` | 10 |
| `time` | 1 |
| `domain provision` | 1 |
| `testparm` | 1 |
| **TOTAL** | **~99 operations** |

---

## 3. Execution Patterns

### 3.1 Two Invocation Methods

The plugin uses **two Cockpit API calls** to execute samba-tool:

#### Method A: `cockpit.script()` — String-based (91 calls, ~90% of code)

Used for simple commands where arguments are interpolated directly into a template string:

```javascript
const command = `samba-tool gpo create ${displayName}`;
const script = () => cockpit.script(command, { superuser: true, err: 'message' })
        .done((data) => {
            setSuccessMessage(data);
            setSuccessAlertVisible(true);
            setLoading(false);
            setIsModalOpen(false);
        })
        .catch((exception) => {
            setErrorMessage(exception.message);
            setErrorAlertVisible(true);
            setLoading(false);
            setIsModalOpen(false);
        });
script();
```

#### Method B: `cockpit.spawn()` — Array-based (10 calls, ~10% of code)

Used for commands with many optional flags (dsacl set, ntacl get/set/changedomsid/sysvolcheck/sysvolreset/getdosinfo, domain demote):

```javascript
const cmd = ["samba-tool", "ntacl", "get", `${file}`];
if (xattrBackend.length > 0) {
    cmd.push(`--xattr-backend=${xattrBackend}`);
}
if (eadbFile.length > 0) {
    cmd.push(`--eadb-file=${eadbFile}`);
}
// ... more optional flags
const script = () => cockpit.spawn(cmd, { superuser: true, err: 'message' })
        .done((data) => { ... })
        .catch((exception) => { ... });
```

### 3.2 Credential Handling

**There is no centralized credential management.** This is one of the project's biggest weaknesses:

1. **Superuser privilege:** Every command passes `{ superuser: true }` — Cockpit handles privilege escalation via PolicyKit/polkit. No `sudo` is hardcoded in command strings.

2. **Per-screen credentials:** DNS operations (`dns/create.js`, `dns/delete.js`, etc.) require `--password=${password}` on every single operation. The user must type the admin password into a TextInput field on each DNS form.

3. **Provisioning:** The `--adminpass` is captured in the provisioning modal and passed directly as a command-line argument (visible in process list — a security concern).

4. **FSMO seize/transfer:** Require `userName` and `adminPass` state variables, passed as command arguments.

5. **No Kerberos, no cached credentials:** The plugin does not use `kinit`, does not cache tickets, does not use `--authentication=kerberos`. All auth is via `--password=` flags or relies on the Cockpit session's local root/sudo privileges.

6. **Fork note:** The `lexandree7/samba-ad-dc` fork specifically states it "removes the old pattern of asking for a DNS server and password on each screen."

### 3.3 Output Parsing

**Extremely naive — the plugin does not parse samba-tool output structurally.** This was flagged by the mentor (Bokovoy) at SambaXP 2021 as a fundamental problem:

1. **Line splitting only:** The universal pattern is:
   ```javascript
   const splitData = data.split('\n');
   setSuccessMessage(splitData);
   ```
   Output is split by newline into an array of strings, then rendered as a `<li>` list or `<h6>` lines.

2. **No structured parsing:** No JSON output (`--json`), no regex extraction, no key-value parsing. The raw text lines from samba-tool are displayed verbatim.

3. **List rendering:** For list commands (`gpo listall`, `computer list`, `group list`, `user list`), the newline-split array is rendered as:
   ```jsx
   {gpoList.map(gpo => <li key={gpo.toString()}>{gpo}</li>)}
   ```

4. **Modal display:** For show/query commands, output is displayed in a Modal popup:
   ```jsx
   <div>{successMessage.map((line) => <h6 key={line.toString()}>{line}</h6>)}</div>
   ```

5. **No filtering on server-side:** Some list views (computer/list.js, group/listgroups.js) implement client-side search filtering:
   ```javascript
   const filteredList = computerList.filter((computer) => computer.includes(searchValue))
   ```

### 3.4 Error Handling

Consistent but basic pattern across all modules:

```javascript
.catch((exception) => {
    setErrorMessage(exception.message);
    setErrorAlertVisible(true);
    setLoading(false);
    setIsModalOpen(false);
});
```

Errors are displayed as toast notifications (`ErrorToast` component) with the raw `exception.message` from Cockpit.

---

## 4. Provisioning Wizard

File: `src/provision-modal.js`

The provisioning wizard is a **PatternFly Modal** with a form collecting:

| Field | State Variable | Options |
|---|---|---|
| Realm | `realm` | Free text (e.g., `SAMBA.TEST`) |
| Domain | `domain` | Free text (e.g., `SAMBA`) |
| Server Role | `serverRole` | `dc`, `member`, `standalone` (FormSelect) |
| DNS Backend | `dnsBackend` | `SAMBA_INTERNAL`, `BIND9_FLATFILE`, `BIND9_DLZ`, `NONE` (FormSelect) |
| DNS Forwarder | `dnsforwarder` | Free text (collected but NOT used in command) |
| Admin Password | `password1` / `password2` | Dual field with matching validation |

### Password Validation
```javascript
const passwordChecker = (pass1, pass2) => {
    if (pass1.length === 0) {
        setHelperTextInvalid('Password must be greater than one character');
    } else if (pass1.length < 8) {
        setHelperTextInvalid('Password length must be longer than 8 characters');
    } else if (pass1 != pass2) {
        setHelperTextInvalid("Passwords don't match");
    } else {
        setIsValidPassword(true);
        setIsValidated('success');
    }
};
```

### Provisioning Command (Multi-line Shell Script)
```javascript
const command = `
    samba-tool domain provision \
            --use-rfc2307 \
            --realm ${realm} \
            --domain ${domain} \
            --server-role ${serverRole} \
            --dns-backend ${dnsBackend} \
            --adminpass ${password1}

    cp -f /var/lib/samba/private/krb5.conf /etc/krb5.conf
    samba
    samba-tool domain info 127.0.0.1
`;
```

This is a single `cockpit.script()` call that runs a multi-line shell script: provision → copy krb5.conf → start samba daemon → query domain info. Notably:
- `--use-rfc2307` is hardcoded (always enabled)
- `dnsforwarder` is collected but never used
- The script does not check if samba is already running
- No error recovery if provision succeeds but krb5.conf copy fails

### Entry Gate (`ad-dc-status.js`)
The app first checks if the server is already an AD DC:
```javascript
const command = 'samba-tool testparm --parameter-name=serverrole';
cockpit.script(command, { superuser: true, err: "message" })
    .then((data) => {
        if (data.includes("active directory domain controller")) {
            setAdDcStatus(true);  // Show main dashboard
        } else {
            setAdDcStatus(false); // Show provisioning wizard
        }
    })
```

---

## 5. GPO Management

14 operations covering `samba-tool gpo` comprehensively. This is the most complete module.

### Operations:
1. **Create** — `gpo create ${displayName}` → Modal with TextInput
2. **Delete** — `gpo del ${gpo}` → Modal with GPO ID/GUID input
3. **Backup** — `gpo backup ${gpo}` → Modal with GPO ID input
4. **Restore** — `gpo restore ${displayName} ${location}` → Modal with display name + location
5. **Delete Link** — `gpo dellink ${container} ${gpo}` → Modal with container DN + GPO ID
6. **Set Link** — `gpo setlink ${container} ${gpo}` → Modal with container DN + GPO ID
7. **Get Link** — `gpo getlink ${container}` → Modal with container DN, output in results Modal
8. **Get Inheritance** — `gpo getinheritance ${container}` → Modal with container DN
9. **Set Inheritance** — `gpo setinheritance ${container} ${choice}` → Modal with container + choice
10. **List All** — `gpo listall` → Auto-load on mount (`useEffect`), Card with `<li>` list
11. **List for Account** — `gpo list ${account}` → Modal with account name
12. **List Containers** — `gpo listcontainers ${gpo}` → Modal with GPO ID
13. **Fetch** — `gpo fetch ${gpo}` → Modal with GPO ID
14. **Show** — `gpo show ${gpo}` → Modal with GPO ID, output displayed in results Modal

### UI Pattern (GPO):
All GPO operations follow one of two patterns:

**Action Pattern (create/delete/backup/set):** Button → Modal form → `cockpit.script` → SuccessToast/ErrorToast

**Query Pattern (show/getlink/listall):** Button → Modal form (or auto-load) → Results displayed in separate Modal with line-by-line `<h6>` rendering

---

## 6. DS ACLs and NT ACLs

### DS ACLs (`samba-tool dsacl`)

Two operations, both using `cockpit.spawn()` with array-based arguments:

**Get (`dsacl/get.js`):**
```javascript
const command = `samba-tool dsacl get`;
cockpit.script(command, { superuser: true, err: 'message' })
    .done((data) => {
        const splitData = data.split('\n');
        setaccessList(splitData);
    })
```
Auto-loads on mount via `useEffect`. Displays in a Card with `<li>` list.

**Set (`dsacl/set.js`):**
```javascript
const cmd = ["samba-tool", "dsacl", "set"];
if (url.length > 0) cmd.push(`--URL=${url}`);
if (car.length > 0) cmd.push(`--car=${car}`);
if (action.length > 0) cmd.push(`--action=${action}`);
if (objectdn.length > 0) cmd.push(`--objectdn=${objectdn}`);
if (trusteedn.length > 0) cmd.push(`--trusteedn=${trusteedn}`);
if (sddl.length > 0) cmd.push(`--sddl=${sddl}`);
cockpit.spawn(cmd, { superuser: true, err: 'message' })
```
Modal with 6 TextInput fields for the optional flags. Uses `cockpit.spawn` (array form) — more robust than template strings.

### NT ACLs (`samba-tool ntacl`)

Six operations, all using `cockpit.spawn()` with array-based arguments and optional flags:

| Operation | Command | Required Args | Optional Flags |
|---|---|---|---|
| `get` | `ntacl get ${file}` | file path | `--xattr-backend`, `--eadb-file`, `--use-ntvfs`, `--use-s3fs`, `--service` |
| `set` | `ntacl set ${acl} ${file}` | ACL string, file path | same as above |
| `changedomsid` | `ntacl changedomsid ${origSid} ${newSid} ${file}` | orig SID, new SID, file | same as above |
| `getdosinfo` | `ntacl getdosinfo ${file}` | file path | same as above |
| `sysvolcheck` | `ntacl sysvolcheck ${file}` | file path | — |
| `sysvolreset` | `ntacl sysvolreset ${file}` | file path | — |

All NT ACL operations use the `cockpit.spawn()` array pattern:
```javascript
const cmd = ["samba-tool", "ntacl", "get", `${file}`];
if (xattrBackend.length > 0) cmd.push(`--xattr-backend=${xattrBackend}`);
if (eadbFile.length > 0) cmd.push(`--eadb-file=${eadbFile}`);
if (useNtvfs.length > 0) cmd.push(`--use-ntvfs=${useNtvfs}`);
if (useS3fs.length > 0) cmd.push(`--use-s3fs=${useS3fs}`);
if (service.length > 0) cmd.push(`--service=${service}`);
cockpit.spawn(cmd, { superuser: true, err: 'message' })
```

---

## 7. FSMO Management

Three operations:

**Show (`fsmo/show.js`):**
```javascript
const command = `samba-tool fsmo show`;
cockpit.script(command, { superuser: true, err: 'message' })
    .done((data) => {
        const splitData = data.split('\n');
        setRoles(splitData);
    })
```
Auto-loads on mount. Displays roles in a Card with `<li>` list.

**Seize (`fsmo/seize.js`):**
Modal with:
- `userName` — Admin username
- `adminPass` — Admin password
- 8 Checkbox toggles for FSMO roles: `rid`, `schema`, `pdc`, `naming`, `infrastructure`, `domainDns`, `forestDns`, `all`
- Comments document the role mappings: `rid=RidAllocationMasterRole`, `schema=SchemaMasterRole`, `pdc=PdcEmulationMasterRole`, etc.

The checkboxes are used to build `--role=` flags for the command.

**Transfer (`fsmo/transfer.js`):**
Identical UI structure to seize — same 8 checkboxes, same admin credentials fields. Only the samba-tool subcommand differs (`fsmo transfer` vs `fsmo seize`).

---

## 8. Trusts

Six trust operations under `domain/trust/`:

| Operation | File | Command | UI Pattern |
|---|---|---|---|
| Create | `trust/create.js` | `samba-tool domain trust create ${domain}` | Modal with domain name → SuccessToast |
| Delete | `trust/delete.js` | `samba-tool domain trust delete ${domain}` | Modal with domain name → SuccessToast |
| List | `trust/list.js` | `samba-tool domain trust list` | Button → loading Modal → results Modal with list |
| Show | `trust/show.js` | `samba-tool domain trust show ${name}` | Modal with trust name → results Modal |
| Validate | `trust/validate.js` | `samba-tool domain trust validate ${domain}` | Modal with domain → results Modal |
| Namespaces | `trust/namespaces.js` | `samba-tool domain namespaces online ${domain}` | Modal with domain → results Modal |

All trust operations follow the standard pattern: Button → input Modal → `cockpit.script()` → output in results Modal (line-by-line `<h6>`) or toast notification.

**Note:** The `namespaces` command uses `samba-tool domain namespaces online` which may be incorrect — the correct samba-tool subcommand for forest trust namespace management is `samba-tool domain trust namespaces`.

---

## 9. UI Patterns

### 9.1 Navigation Architecture

The plugin uses **multi-page HTML navigation**, not SPA routing:

- `index.html` is the main dashboard
- Each module has its own `<module>.html` (e.g., `computer/computer.html`, `gpo/gpo.html`)
- Navigation is via `<a href>` links — clicking a card navigates to a new HTML page
- Each HTML page loads its own JS bundle via webpack
- A `BackButton` component uses `history.back()` for navigation

```jsx
// main.js — Dashboard cards
<Gallery hasGutter>
    <GalleryItem>
        <a href="computer/computer.html" role="link">
            <Card isHoverable><CardBody>Computer Management</CardBody></Card>
        </a>
    </GalleryItem>
    ...
</Gallery>
```

### 9.2 Component Patterns

**Three repeating component archetypes:**

1. **List View** (auto-load on mount):
   ```jsx
   useEffect(() => {
       setLoading(true);
       const command = `samba-tool <subcmd> list`;
       cockpit.script(command, { superuser: true, err: 'message' })
           .done((data) => {
               const splitData = data.split('\n');
               const sortedData = splitData.sort();
               setList(sortedData);
               setLoading(false);
           })
   }, []);
   // Render: Card → CardBody → Loading + RenderError + <li> list
   ```

2. **Action Modal** (create/delete/set):
   ```jsx
   // Button triggers Modal with form
   <Button variant="primary" onClick={handleModalToggle}>Create GPO</Button>
   <Modal title="..." isOpen={isModalOpen} onClose={handleModalToggle}
       actions={[
           <Button onClick={handleSubmit}>Create</Button>,
           <Button variant="link" onClick={handleModalToggle}>Cancel</Button>
       ]}>
       <Form><FormGroup><TextInput .../></FormGroup></Form>
   </Modal>
   ```

3. **Query Results Modal** (show/get/info):
   ```jsx
   // Button triggers input Modal, results shown in separate Modal
   <Modal isOpen={successAlertVisible} onClose={...}>
       <div>{successMessage.map((line) => <h6>{line}</h6>)}</div>
   </Modal>
   ```

### 9.3 State Management

Every component uses React `useState` hooks. The state variables are highly repetitive:
- `loading`, `errorMessage`, `errorAlertVisible`, `successMessage`, `successAlertVisible`, `isModalOpen`
- Input fields as individual state variables

No global state, no context, no Redux. Each page is self-contained.

### 9.4 Shared Components (`common.js`)

| Component | Purpose |
|---|---|
| `RenderError` | Inline Alert (danger) with close button |
| `Success` | Inline Alert (success) with close button |
| `Loading` | Spinner wrapper (shows/hides based on `loading` prop) |
| `BackButton` | PatternFly tertiary button with AngleLeftIcon, calls `history.back()` |
| `ErrorToast` | Toast AlertGroup (danger) for transient error notifications |
| `SuccessToast` | Toast AlertGroup (success) for transient success notifications |

### 9.5 CSS

- `css/common.css` — shared styles
- `app.scss` — root styles
- `lib/patternfly-4-cockpit.scss` — PatternFly imports
- `lib/patternfly-4-overrides.scss` — Cockpit-specific overrides
- Module-specific `index.css` files (sites, user, domain)

---

## 10. Reusable Snippets (samba-tool Wrappers)

### Snippet 1: Simple Command Wrapper (template string)
```javascript
import cockpit from 'cockpit';

function runSambaTool(command, onSuccess, onError) {
    cockpit.script(command, { superuser: true, err: 'message' })
        .done((data) => onSuccess(data))
        .catch((exception) => onError(exception.message));
}

// Usage:
runSambaTool(
    `samba-tool user list`,
    (data) => console.log(data.split('\n')),
    (err) => console.error(err)
);
```

### Snippet 2: Command with Optional Flags (array-based, safe)
```javascript
import cockpit from 'cockpit';

function runSambaToolArray(subcmd, args = [], flags = {}) {
    const cmd = ['samba-tool', ...subcmd.split(' '), ...args];
    for (const [key, value] of Object.entries(flags)) {
        if (value && value.length > 0) {
            cmd.push(`--${key}=${value}`);
        }
    }
    return cockpit.spawn(cmd, { superuser: true, err: 'message' });
}

// Usage:
runSambaToolArray('ntacl get', ['/path/to/file'], {
    'xattr-backend': 'xattr_tdb',
    'use-s3fs': 'yes'
}).done(console.log).catch(err => console.error(err));
```

### Snippet 3: Auto-Load List Component (React Hook Pattern)
```javascript
import React, { useState, useEffect } from 'react';
import cockpit from 'cockpit';
import { Loading, RenderError } from './common';
import { Card, CardBody, CardHeader } from '@patternfly/react-core';

function SambaToolList({ subcmd, title }) {
    const [items, setItems] = useState([]);
    const [error, setError] = useState();
    const [loading, setLoading] = useState(false);
    const [alertVisible, setAlertVisible] = useState(false);

    useEffect(() => {
        setLoading(true);
        cockpit.script(`samba-tool ${subcmd}`, { superuser: true, err: 'message' })
            .done((data) => {
                setItems(data.split('\n').filter(Boolean).sort());
                setLoading(false);
            })
            .catch((exception) => {
                setError(exception.message);
                setAlertVisible(true);
                setLoading(false);
            });
    }, []);

    return (
        <Card>
            <CardHeader>{title}</CardHeader>
            <CardBody>
                {loading && <Spinner />}
                {alertVisible && <RenderError error={error} />}
                {items.map(item => <li key={item}>{item}</li>)}
            </CardBody>
        </Card>
    );
}
```

### Snippet 4: Action Modal Wrapper (create/delete/set)
```javascript
import React, { useState } from 'react';
import cockpit from 'cockpit';
import { Modal, Button, Form, FormGroup, TextInput } from '@patternfly/react-core';
import { Loading, SuccessToast, ErrorToast } from './common';

function SambaToolAction({ buttonLabel, modalTitle, commandTemplate, fields }) {
    const [isModalOpen, setIsModalOpen] = useState(false);
    const [loading, setLoading] = useState(false);
    const [values, setValues] = useState({});
    const [successMsg, setSuccessMsg] = useState('');
    const [errorMsg, setErrorMsg] = useState('');
    const [showSuccess, setShowSuccess] = useState(false);
    const [showError, setShowError] = useState(false);

    const handleSubmit = () => {
        setLoading(true);
        const command = commandTemplate(values);
        cockpit.script(command, { superuser: true, err: 'message' })
            .done((data) => {
                setSuccessMsg(data);
                setShowSuccess(true);
                setLoading(false);
                setIsModalOpen(false);
            })
            .catch((exception) => {
                setErrorMsg(exception.message);
                setShowError(true);
                setLoading(false);
                setIsModalOpen(false);
            });
    };

    return (
        <>
            {showError && <ErrorToast errorMessage={errorMsg} closeModal={() => setShowError(false)} />}
            {showSuccess && <SuccessToast successMessage={successMsg} closeModal={() => setShowSuccess(false)} />}
            <Button variant="primary" onClick={() => setIsModalOpen(true)}>{buttonLabel}</Button>
            <Modal title={modalTitle} isOpen={isModalOpen} onClose={() => setIsModalOpen(false)}
                actions={[
                    <Button key="confirm" variant="primary" onClick={handleSubmit}>Confirm</Button>,
                    <Button key="cancel" variant="link" onClick={() => setIsModalOpen(false)}>Cancel</Button>
                ]}>
                <Form>
                    {fields.map(f => (
                        <FormGroup key={f.name} label={f.label} isRequired={f.required}>
                            <TextInput value={values[f.name] || ''}
                                onChange={v => setValues({...values, [f.name]: v})} />
                        </FormGroup>
                    ))}
                </Form>
            </Modal>
        </>
    );
}
```

### Snippet 5: Query Results Modal (show/get/info)
```javascript
function SambaToolQuery({ buttonLabel, modalTitle, commandTemplate, inputFields }) {
    const [isModalOpen, setIsModalOpen] = useState(false);
    const [resultsModalOpen, setResultsModalOpen] = useState(false);
    const [loading, setLoading] = useState(false);
    const [results, setResults] = useState([]);
    const [values, setValues] = useState({});

    const handleSubmit = () => {
        setLoading(true);
        const command = commandTemplate(values);
        cockpit.script(command, { superuser: true, err: 'message' })
            .done((data) => {
                setResults(data.split('\n'));
                setResultsModalOpen(true);
                setLoading(false);
                setIsModalOpen(false);
            })
            .catch((err) => { console.error(err); setLoading(false); });
    };

    return (
        <>
            <Button variant="secondary" onClick={() => setIsModalOpen(true)}>{buttonLabel}</Button>
            <Modal title={modalTitle} isOpen={isModalOpen} onClose={() => setIsModalOpen(false)}>
                {/* input fields */}
            </Modal>
            <Modal title="Results" isOpen={resultsModalOpen} onClose={() => setResultsModalOpen(false)}>
                <div>{results.map((line, i) => <h6 key={i}>{line}</h6>)}</div>
            </Modal>
        </>
    );
}
```

### Snippet 6: Provisioning Command (Multi-step Shell)
```javascript
const provisionCommand = (realm, domain, serverRole, dnsBackend, adminPass) => `
    samba-tool domain provision \
        --use-rfc2307 \
        --realm ${realm} \
        --domain ${domain} \
        --server-role ${serverRole} \
        --dns-backend ${dnsBackend} \
        --adminpass ${adminPass}
    cp -f /var/lib/samba/private/krb5.conf /etc/krb5.conf
    samba
    samba-tool domain info 127.0.0.1
`;
cockpit.script(provisionCommand(realm, domain, role, dns, pass), { superuser: true, err: 'message' });
```

### Snippet 7: DNS with Credentials (Per-Call Password)
```javascript
const dnsAddCmd = (server, zone, name, type, data, password) =>
    `samba-tool dns add ${server} ${zone} ${name} ${type} ${data} --password=${password}`;
cockpit.script(dnsAddCmd(server, zone, name, type, data, pass), { superuser: true, err: 'message' });
```

### Snippet 8: Server Role Check (Gate Function)
```javascript
function checkServerRole() {
    return cockpit.script('samba-tool testparm --parameter-name=serverrole',
        { superuser: true, err: 'message' })
        .then((data) => data.includes('active directory domain controller'));
}
```

### Snippet 9: FSMO Role Selection (Checkbox → Flag Builder)
```javascript
const fsmoRoles = [
    { key: 'rid',       flag: 'rid',          label: 'RID Allocation Master' },
    { key: 'schema',    flag: 'schema',       label: 'Schema Master' },
    { key: 'pdc',       flag: 'pdc',          label: 'PDC Emulator' },
    { key: 'naming',    flag: 'naming',       label: 'Naming Master' },
    { key: 'infra',     flag: 'infrastructure', label: 'Infrastructure Master' },
    { key: 'domainDns', flag: 'domaindns',    label: 'Domain DNS Master' },
    { key: 'forestDns', flag: 'forestdns',    label: 'Forest DNS Master' },
];

function buildFsmoCommand(action, roles, user, pass) {
    let cmd = `samba-tool fsmo ${action}`;
    const selected = roles.filter(r => r.selected);
    if (selected.length === 0) return null;
    selected.forEach(r => { cmd += ` --role=${r.flag}`; });
    cmd += ` --username=${user} --password=${pass}`;
    return cmd;
}
```

### Snippet 10: Client-Side Filtered List (Search)
```javascript
const [searchValue, setSearchValue] = useState('');
const [items, setItems] = useState([]);

const filteredList = items
    .filter(item => item.includes(searchValue))
    .map(item => <li key={item}>{item}</li>);

// In JSX:
<InputGroup>
    <TextInput value={searchValue} onChange={setSearchValue} placeholder="Search..." />
    <Button variant={ButtonVariant.tertiary}><SearchIcon /></Button>
</InputGroup>
{filteredList}
```

---

## 11. Why It Stalled

Multiple converging factors:

### 11.1 Architectural Deficits

1. **No structured output parsing.** The plugin splits samba-tool's free-text output by `\n` and renders raw lines. As Bokovoy noted at SambaXP 2021: *"Parsing human-oriented output is a waste of resources for robots."* Samba-tool lacks stable machine-readable output (no `--json` flag), making any UI built on text parsing inherently fragile — any samba-tool output format change breaks the UI.

2. **No credential abstraction.** Every DNS operation requires a password TextInput. FSMO operations require username + password. There is no Kerberos ticket caching, no session-based auth, no centralized credential store. The user re-enters credentials on every screen.

3. **Shell injection risk.** Template string interpolation (``${command} ${userInput}``) passes unsanitized user input directly to the shell. While Cockpit runs as the authenticated user, this is still a vulnerability if any input contains shell metacharacters.

4. **No error recovery or state validation.** The provisioning wizard runs a multi-line shell script with no intermediate error checking. If `samba-tool domain provision` succeeds but `cp` or `samba` fails, the wizard shows success.

### 11.2 Project Factors

5. **GSoC project ended.** The last commit is August 28, 2020 — the GSoC coding period ended. Hezekiah was a Real Estate student new to Linux and Samba; he did not continue contributing after GSoC.

6. **No maintainer adoption.** The Samba Team (Bokovoy) hosted it on GitLab under `samba-team/cockpit-samba-ad-dc` but no one took over maintenance. The project remained a "prototype" (Bokovoy's words at SambaXP 2021).

7. **Moving target.** Samba evolves rapidly. samba-tool subcommands, flags, and output formats change between releases. Without active maintenance, the plugin becomes incompatible with newer Samba versions. The fork by `lexandree7` attempted to modernize for Samba 4.15.13 but also appears inactive.

8. **No upstream integration.** The plugin was never merged into Cockpit itself or into Samba's main repository. It lived as a standalone OBS package that depended on specific Samba versions.

9. **No testing.** The `test/` directory exists but tests are minimal. No integration tests against a real Samba AD DC. CI was Travis-based but did not exercise samba-tool commands.

### 11.3 The Fundamental Problem (from Bokovoy's SambaXP 2021 talk)

The core issue identified by the mentor: **samba-tool produces human-oriented output, not machine-oriented output.** Building a reliable UI on top of free-text parsing is fundamentally fragile. Bokovoy's conclusion: *"We can do better (for robots and humans). A little magic can help both"* — suggesting that Samba needs a proper machine-readable API (JSON output, DBus interface, or Python library bindings) rather than wrapping CLI text output.

This insight is directly relevant to SambaForge: **any UI wrapping samba-tool must handle the absence of structured output** — either by parsing carefully with knowledge of samba-tool's format, by contributing `--json` flags upstream, or by using Samba's Python bindings directly instead of shelling out.

---

## Key Takeaways for SambaForge

1. **Highest samba-tool coverage** of any known project: ~99 operations across 17 subcommand groups
2. **cockpit.script + cockpit.spawn** is the execution model — no DBus, no Python bridge
3. **superuser: true** via Cockpit/PolicyKit handles privilege escalation cleanly
4. **Naive text parsing** (`split('\n')`) is the Achilles heel — needs structured output
5. **Per-screen credential entry** is a UX disaster — needs centralized auth
6. **Template string commands** are injection-vulnerable — use array-based `spawn()` instead
7. **Multi-page HTML navigation** (not SPA) — each module is a separate Cockpit page
8. **PatternFly 4 + React hooks** is the UI stack — clean but highly repetitive code
9. **Provisioning wizard** is a single multi-line shell script with no error recovery
10. **The project stalled because** GSoC ended, no maintainer adopted it, samba-tool output is unstable for parsing, and there's no credential management layer