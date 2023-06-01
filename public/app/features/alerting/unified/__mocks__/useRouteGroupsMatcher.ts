import { useCallback } from 'react';

import { Labels } from '@grafana/data';

import { AlertmanagerGroup, RouteWithID } from '../../../../plugins/datasource/alertmanager/types';
import { findMatchingRoutes, normalizeRoute, RouteInstanceMatch } from '../utils/notification-policies';

export function useRouteGroupsMatcher() {
  const getRouteGroupsMap = useCallback(async (route: RouteWithID, __: AlertmanagerGroup[]) => {
    const groupsMap = new Map<string, AlertmanagerGroup[]>();
    function addRoutes(route: RouteWithID) {
      groupsMap.set(route.id, []);

      route.routes?.forEach((r) => addRoutes(r));
    }

    addRoutes(route);

    return groupsMap;
  }, []);

  const matchInstancesToRoute = useCallback(async (rootRoute: RouteWithID, instancesToMatch: Labels[]) => {
    const result = new Map<string, RouteInstanceMatch[]>();

    const normalizedRootRoute = normalizeRoute(rootRoute);

    instancesToMatch.forEach((instance) => {
      const matchingRoutes = findMatchingRoutes(normalizedRootRoute, Object.entries(instance));
      matchingRoutes.forEach(({ route, details }) => {
        // Only to convert Label[] to Labels[] - needs better approach
        const matchDetails = new Map(
          Array.from(details.entries()).map(([matcher, labels]) => [matcher, Object.fromEntries(labels)])
        );

        const currentRoute = result.get(route.id);
        if (currentRoute) {
          currentRoute.push({ route, labels: instance, matchDetails });
        } else {
          result.set(route.id, [{ route, labels: instance, matchDetails }]);
        }
      });
    });

    return result;
  }, []);

  return { getRouteGroupsMap, matchInstancesToRoute };
}
