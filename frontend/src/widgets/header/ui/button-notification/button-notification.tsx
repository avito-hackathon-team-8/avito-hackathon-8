import clsx from "clsx";

import { NotificationIcon } from "@/shared/assets/icon/notification";
import { BottomPanel } from "@/shared/ui/bottom-panel";
import { Button } from "@/shared/ui/button";
import { Typography } from "@/shared/ui/typography";

import styles from "./button-notification.module.scss";
interface IButtonNotificationProps {
  className?: string;
}

export const ButtonNotification = ({ className }: IButtonNotificationProps) => {
  return (
    <BottomPanel
      closeOnBackdrop
      title="Уведомления"
      renderTrigger={(open) => (
        <Button
          className={clsx(styles.notificationButton, className)}
          aria-label="Открыть уведомления"
          onClick={open}
        >
          <>
            <NotificationIcon />
            <Typography
              className={styles.notification__count}
              variant="caption"
              as="span"
              color="white"
              aria-label="Кол-во уведомлений"
            >
              3
            </Typography>
          </>
        </Button>
      )}
    >
      <div>
        Lorem ipsum dolor sit amet consectetur, adipisicing elit. In vitae eius
        corrupti beatae, veritatis aliquam omnis vel ipsum distinctio dicta,
        dolor, eaque earum necessitatibus quidem a? Sequi dolorum cumque vitae?
      </div>
    </BottomPanel>
  );
};
