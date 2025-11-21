import { Modal, Tag, Typography } from "antd";
import React from "react";
import styles from "./PlayUrlModal.less";

interface PlayUrlModalProps {
  open: boolean;
  path?: string;
  onClose: () => void;
}

const PlayUrlModal: React.FC<PlayUrlModalProps> = ({ open, path, onClose }) => {
  return (
    <Modal
      title="播放地址"
      open={open}
      onCancel={onClose}
      footer={null}
      width={960}
      destroyOnHidden
    >
      {path ? (
        <div>
          <div className={styles.headerRow}>
            <div className={styles.headerProtocol}>协议</div>
            <div className={styles.headerUrl}>播放地址</div>
          </div>
          {(() => {
            const address = window.location.host;
            const protocol =
              window.location.protocol === "https:" ? "https" : "http";
            const wsProtocol =
              window.location.protocol === "https:" ? "wss" : "ws";
            const hls = `${protocol}://${address}/hls/${path}.m3u8`;
            const ws = `${wsProtocol}://${address}/flv/${path}`;
            const rtsp = `rtsp://${address}/${path}`;
            const Row: React.FC<{ name: string; url: string }> = ({
              name,
              url,
            }) => (
              <div className={styles.dataRow}>
                <div className={styles.protocolCell}>
                  <Tag>{name}</Tag>
                </div>
                <div className={styles.urlCell}>
                  <Typography.Paragraph
                    copyable={{ text: url }}
                    className={styles.urlText}
                  >
                    {url}
                  </Typography.Paragraph>
                </div>
              </div>
            );
            return (
              <>
                <Row name="hls" url={hls} />
                <Row name="ws-flv" url={ws} />
                <Row name="rtsp" url={rtsp} />
              </>
            );
          })()}
        </div>
      ) : null}
    </Modal>
  );
};

export default PlayUrlModal;
