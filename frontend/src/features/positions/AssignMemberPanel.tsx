import { EntityCreateSurface } from "../../shared/ui/surfaces";
import { AssignMemberForm } from "./AssignMemberForm";

type AssignMemberPanelProps = {
  positionId: string;
  positionTitle: string;
  onSuccess: () => void;
  onClose: () => void;
};

export function AssignMemberPanel({ positionId, positionTitle, onSuccess, onClose }: AssignMemberPanelProps) {
  return (
    <EntityCreateSurface showTitle={false} onClose={onClose}>
      <AssignMemberForm
        positionId={positionId}
        positionTitle={positionTitle}
        onSuccess={onSuccess}
        onCancel={onClose}
      />
    </EntityCreateSurface>
  );
}
