# Nota técnica: Samba AD DC en Proxmox LXC

## Problema

La VM de prueba (172.30.36.91) es un **contenedor LXC** de Proxmox, no una VM completa (KVM). 

### Limitaciones detectadas

1. **xattrs del sistema**: LXC no soporta `system.*` xattrs que Samba necesita para almacenar ACLs NT. Workaround: usar `xattr_tdb` para almacenar xattrs en un archivo TDB en vez del filesystem.

2. **samba-tool domain provision PANIC**: Sin `xattr_tdb`, `samba-tool domain provision` crashea con "Security context active token stack underflow" en `acl_xattr.so`. Con `--option='vfs objects = acl_xattr xattr_tdb'` el provisioning completa pero toma >5 minutos.

3. **Samba AD DC no arranca**: Después del provisioning, `samba-ad-dc` falla con "Failed to obtain server credentials, perhaps a standalone server?: NT_STATUS_CANT_ACCESS_DOMAIN_INFO". Esto es porque LXC no permite que el proceso Samba obtenga las credenciales de la máquina del keytab del kernel.

4. **Chrony no funciona**: "adjtimex(0x8001) failed: Operation not permitted" — LXC no permite operaciones de reloj del kernel.

## Solución

**Para desarrollo y testing de SambaForge:** Usar una **VM KVM completa** en Proxmox, no un contenedor LXC.

En Proxmox, crear la VM con:
- Tipo: **Virtual Machine (KVM)**, no Container (LXC)
- OS: Debian 13 minimal
- BIOS: OVMF (UEFI) o SeaBIOS
- Disk: virtio-scsi
- Network: virtio

Esto da un kernel completo con soporte de xattrs del sistema, operaciones de reloj, y todas las syscalls que Samba AD DC necesita.

## Lección para SambaForge

El script de instalación (`install.sh`) debe detectar si corre en un contenedor LXC y advertir al usuario:

```bash
if systemd-detect-virt | grep -q lxc; then
    echo "ADVERTENCIA: Samba AD DC no funciona correctamente en contenedores LXC."
    echo "Use una VM completa (KVM) en Proxmox."
    # No abortar — el usuario puede querer probar de todas formas
fi
```

Esto se añadirá como check P-19 en el ADR-010.