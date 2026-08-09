import { BottomPanel } from '@/shared/ui/bottom-panel';

import { useActivityDaysAutoPanel } from '../../model/use-activity-days-auto-panel';
import { ActivityDaysPanelContent } from '../activity-days-panel-content/activity-days-panel-content';

const TITLE = 'Дни активности';
const DESCRIPTION = 'Награда зависит от календарного дня';

export const ActivityDaysAutoPanel = () => {
  const { data, isOpen, handleReceiveReward } = useActivityDaysAutoPanel();

  if (!isOpen) return null;
  return (
    <BottomPanel
      title={TITLE}
      description={DESCRIPTION}
      isStartOpen={isOpen}
      renderTrigger={() => null}
    >
      {data && <ActivityDaysPanelContent data={data} handleReceiveReward={handleReceiveReward} />}
    </BottomPanel>
  );
};
